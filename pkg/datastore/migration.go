package datastore

import (
	"fmt"
	"sort"
	"time"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/errs"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/settings"

	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// Migration is a single, ordered schema or data change.
//
// Up and the record that marks the migration as applied are committed together,
// so a failed migration leaves nothing behind and can simply be retried.
type Migration struct {
	// Id is unique and sorts chronologically, e.g. 20261030_01_shared_expenses.
	Id string
	// Description is written to the log when the migration is applied.
	Description string
	// Up applies the change. Every statement has to go through tx, which is the
	// migration's transaction; c may be nil when no request context is available.
	Up func(c core.Context, tx *MigrationTx) error
}

// MigrationTx is the handle a migration uses to apply its change. Opening a new
// session would run outside the migration transaction, so use tx for everything.
type MigrationTx struct {
	db      *Database
	session *xorm.Session
}

// Session returns the transaction session, for statements the helpers do not cover
func (tx *MigrationTx) Session() *xorm.Session {
	return tx.session
}

// Exec runs a statement inside the migration transaction
func (tx *MigrationTx) Exec(statement string, args ...any) error {
	statementArgs := make([]any, 0, len(args)+1)
	statementArgs = append(statementArgs, statement)
	statementArgs = append(statementArgs, args...)

	_, err := tx.session.Exec(statementArgs...)

	return err
}

// DatabaseType returns the configured database type, e.g. "sqlite3"
func (tx *MigrationTx) DatabaseType() string {
	return tx.db.databaseType
}

// HasTable reports whether the table exists, as seen from inside the transaction.
// It queries the transaction session rather than the engine, because the engine
// would need a second connection while the transaction holds one, which deadlocks
// when the pool allows a single connection (the SQLite default).
func (tx *MigrationTx) HasTable(tableName string) (bool, error) {
	return tx.exists(
		"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
		"SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name=?",
		"SELECT table_name FROM information_schema.tables WHERE table_schema = current_schema() AND table_name=?",
		tableName,
	)
}

// HasColumn reports whether the column exists on the table, as seen from inside
// the transaction.
func (tx *MigrationTx) HasColumn(tableName string, columnName string) (bool, error) {
	return tx.exists(
		"SELECT name FROM pragma_table_info(?) WHERE name=?",
		"SELECT column_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name=? AND column_name=?",
		"SELECT column_name FROM information_schema.columns WHERE table_schema = current_schema() AND table_name=? AND column_name=?",
		tableName,
		columnName,
	)
}

func (tx *MigrationTx) exists(sqliteStatement string, mysqlStatement string, postgresStatement string, args ...any) (bool, error) {
	var statement string

	switch tx.db.databaseType {
	case settings.Sqlite3DbType:
		statement = sqliteStatement
	case settings.MySqlDbType:
		statement = mysqlStatement
	case settings.PostgresDbType:
		statement = postgresStatement
	default:
		return false, errs.ErrDatabaseTypeInvalid
	}

	rows, err := tx.session.QueryString(append([]any{statement}, args...)...)

	if err != nil {
		return false, err
	}

	return len(rows) > 0, nil
}

// migrations is the ordered registry of all migrations. Append new entries at
// the end; never reorder or rename an existing one.
var migrations = []Migration{
	{
		Id:          "20261030_01_split_direction",
		Description: "default existing transaction splits to \"they owe me\"",
		Up: func(c core.Context, tx *MigrationTx) error {
			// Splits created before directions existed were all shares owed by the
			// other person, which is the zero value's meaning once defaulted.
			return tx.Exec("UPDATE transaction_split SET direction = ? WHERE direction = 0", models.SPLIT_DIRECTION_THEY_OWE_ME)
		},
	},
}

// RunMigrations applies every pending migration in order and returns the ids
// that were applied during this call.
func (c *DataStoreContainer) RunMigrations(ctx core.Context) ([]string, error) {
	database := c.UserDataStore.Get(0)

	if database == nil {
		return nil, errs.ErrDatabaseIsNull
	}

	return runMigrations(ctx, database, migrations)
}

func runMigrations(ctx core.Context, database *Database, all []Migration) ([]string, error) {
	if err := ensureSchemaMigrationTable(database); err != nil {
		return nil, err
	}

	applied, err := appliedMigrationIds(ctx, database)

	if err != nil {
		return nil, err
	}

	ordered := make([]Migration, len(all))
	copy(ordered, all)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Id < ordered[j].Id })

	var appliedNow []string

	for _, migration := range ordered {
		if applied[migration.Id] {
			continue
		}

		err := database.DoTransaction(ctx, func(session *xorm.Session) error {
			if err := migration.Up(ctx, &MigrationTx{db: database, session: session}); err != nil {
				return fmt.Errorf("migration %q failed: %w", migration.Id, err)
			}

			record := &models.SchemaMigration{
				MigrationId:     migration.Id,
				AppliedUnixTime: time.Now().Unix(),
			}

			if _, err := session.Exec("INSERT INTO schema_migration (migration_id, applied_unix_time) VALUES (?, ?)", record.MigrationId, record.AppliedUnixTime); err != nil {
				return fmt.Errorf("failed to record migration %q: %w", migration.Id, err)
			}

			return nil
		})

		if err != nil {
			return appliedNow, err
		}

		appliedNow = append(appliedNow, migration.Id)
	}

	return appliedNow, nil
}

func ensureSchemaMigrationTable(database *Database) error {
	session := database.NewSession(nil)
	defer session.Close()

	return session.Sync2(new(models.SchemaMigration))
}

func appliedMigrationIds(ctx core.Context, database *Database) (map[string]bool, error) {
	session := database.NewSession(ctx)
	defer session.Close()

	var records []*models.SchemaMigration

	if err := session.Find(&records); err != nil {
		return nil, err
	}

	applied := make(map[string]bool, len(records))

	for _, record := range records {
		applied[record.MigrationId] = true
	}

	return applied, nil
}

// HasTable reports whether the table exists, so a migration can be written to
// work both on databases created before and after a change.
func (db *Database) HasTable(tableName string) (bool, error) {
	table, err := db.tableMeta(tableName)

	if err != nil || table == nil {
		return false, err
	}

	return true, nil
}

// HasColumn reports whether the column exists on the table.
func (db *Database) HasColumn(tableName string, columnName string) (bool, error) {
	table, err := db.tableMeta(tableName)

	if err != nil || table == nil {
		return false, err
	}

	return table.GetColumn(columnName) != nil, nil
}

func (db *Database) tableMeta(tableName string) (*schemas.Table, error) {
	tables, err := db.engineGroup.Master().DBMetas()

	if err != nil {
		return nil, err
	}

	for _, table := range tables {
		if table.Name == tableName {
			return table, nil
		}
	}

	return nil, nil
}
