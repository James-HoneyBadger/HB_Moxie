package datastore

import (
	"fmt"
	"strings"

	"xorm.io/xorm"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/errs"
	"github.com/James-HoneyBadger/honey-ledger/pkg/settings"
)

// Database represents a database instance
type Database struct {
	databaseType string
	engineGroup  *xorm.EngineGroup
}

// NewSession starts a new session with the specified context
func (db *Database) NewSession(c core.Context) *xorm.Session {
	return db.engineGroup.Context(NewXOrmContextAdapter(c))
}

// DoTransaction runs a new database transaction
func (db *Database) DoTransaction(c core.Context, fn func(sess *xorm.Session) error) (err error) {
	sess := db.engineGroup.NewSession()

	if c != nil {
		sess.Context(NewXOrmContextAdapter(c))
	}

	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return err
	}

	if err = fn(sess); err != nil {
		_ = sess.Rollback()
		return err
	}

	if err = sess.Commit(); err != nil {
		return err
	}

	return nil
}

// Exec executes a statement with the specified context
func (db *Database) Exec(c core.Context, statement string, args ...any) error {
	sess := db.NewSession(c)
	defer sess.Close()

	statementArgs := make([]any, 0, len(args)+1)
	statementArgs = append(statementArgs, statement)
	statementArgs = append(statementArgs, args...)

	_, err := sess.Exec(statementArgs...)
	return err
}

// VacuumInto writes a consistent snapshot of the SQLite database to the destination path
func (db *Database) VacuumInto(c core.Context, destPath string) error {
	if db.databaseType != settings.Sqlite3DbType {
		return errs.ErrDatabaseTypeInvalid
	}

	escapedPath := strings.ReplaceAll(destPath, "'", "''")
	return db.Exec(c, fmt.Sprintf("VACUUM INTO '%s'", escapedPath))
}

// Close closes the underlying database connections
func (db *Database) Close() error {
	if db == nil || db.engineGroup == nil {
		return nil
	}

	return db.engineGroup.Close()
}

// SetSavePoint sets a save point in the current transaction
func (db *Database) SetSavePoint(sess *xorm.Session, savePointName string) error {
	if db.databaseType == settings.PostgresDbType || db.databaseType == settings.MySqlDbType || db.databaseType == settings.Sqlite3DbType {
		_, err := sess.Exec("SAVEPOINT " + savePointName)
		return err
	}

	return nil
}

// RollbackToSavePoint rolls back to the specified save point in the current transaction
func (db *Database) RollbackToSavePoint(sess *xorm.Session, savePointName string) error {
	if db.databaseType == settings.PostgresDbType || db.databaseType == settings.MySqlDbType || db.databaseType == settings.Sqlite3DbType {
		_, err := sess.Exec("ROLLBACK TO SAVEPOINT " + savePointName)
		return err
	}

	return nil
}
