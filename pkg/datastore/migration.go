package datastore

import (
	"fmt"
	"sort"
	"time"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/errs"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"

	"xorm.io/xorm/schemas"
)

// Migration is a single, ordered schema or data change.
//
// Migrations must be idempotent-friendly: a failure half way through must not
// leave the database in a state the next run cannot handle, because the record
// is only written after Up returns successfully.
type Migration struct {
	// Id is unique and sorts chronologically, e.g. 20261030_01_shared_expenses.
	Id string
	// Description is written to the log when the migration is applied.
	Description string
	// Up applies the change. c may be nil when no request context is available.
	Up func(c core.Context, db *Database) error
}

// migrations is the ordered registry of all migrations. Append new entries at
// the end; never reorder or rename an existing one.
var migrations = []Migration{}

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

		if err := migration.Up(ctx, database); err != nil {
			return appliedNow, fmt.Errorf("migration %q failed: %w", migration.Id, err)
		}

		record := &models.SchemaMigration{
			MigrationId:     migration.Id,
			AppliedUnixTime: time.Now().Unix(),
		}

		if err := database.Exec(ctx, "INSERT INTO schema_migration (migration_id, applied_unix_time) VALUES (?, ?)", record.MigrationId, record.AppliedUnixTime); err != nil {
			return appliedNow, fmt.Errorf("failed to record migration %q: %w", migration.Id, err)
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
