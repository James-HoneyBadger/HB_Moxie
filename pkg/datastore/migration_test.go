package datastore

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/settings"
)

func newMigrationTestDatabase(t *testing.T) *Database {
	t.Helper()

	dbConfig := &settings.DatabaseConfig{
		DatabaseType: settings.Sqlite3DbType,
		DatabasePath: filepath.Join(t.TempDir(), "migration-test.db"),
	}

	database, err := initializeDatabase(dbConfig)
	require.Nil(t, err)

	t.Cleanup(func() {
		_ = database.Close()
	})

	return database
}

func TestRunMigrationsAppliesInOrderAndOnlyOnce(t *testing.T) {
	database := newMigrationTestDatabase(t)

	var applied []string

	all := []Migration{
		{
			Id:          "20260102_02_second",
			Description: "second",
			Up: func(c core.Context, tx *MigrationTx) error {
				applied = append(applied, "second")
				return tx.Exec("CREATE TABLE second_table (id INTEGER PRIMARY KEY)")
			},
		},
		{
			Id:          "20260101_01_first",
			Description: "first",
			Up: func(c core.Context, tx *MigrationTx) error {
				applied = append(applied, "first")
				return tx.Exec("CREATE TABLE first_table (id INTEGER PRIMARY KEY)")
			},
		},
	}

	appliedNow, err := runMigrations(core.NewNullContext(), database, all)
	require.Nil(t, err)

	// Registered out of order, but executed chronologically by id.
	assert.Equal(t, []string{"first", "second"}, applied)
	assert.Equal(t, []string{"20260101_01_first", "20260102_02_second"}, appliedNow)

	hasFirst, err := database.HasTable("first_table")
	require.Nil(t, err)
	assert.True(t, hasFirst)

	// A second run must not apply anything again, which is what makes adding a
	// migration safe for existing installations.
	appliedAgain, err := runMigrations(core.NewNullContext(), database, all)
	require.Nil(t, err)
	assert.Empty(t, appliedAgain)
	assert.Len(t, applied, 2)
}

func TestRunMigrationsStopsOnFailureWithoutRecordingIt(t *testing.T) {
	database := newMigrationTestDatabase(t)

	all := []Migration{
		{
			Id: "20260101_01_broken",
			Up: func(c core.Context, tx *MigrationTx) error {
				return assert.AnError
			},
		},
	}

	appliedNow, err := runMigrations(core.NewNullContext(), database, all)

	assert.NotNil(t, err)
	assert.Empty(t, appliedNow)
	assert.Contains(t, err.Error(), "20260101_01_broken")

	// The failed migration is not recorded, so the next run retries it.
	var records int64

	session := database.NewSession(core.NewNullContext())
	defer session.Close()

	count, err := session.Count(new(models.SchemaMigration))
	require.Nil(t, err)
	records = count
	assert.Equal(t, int64(0), records)
}

func TestRunMigrationsRollsBackAFailedMigration(t *testing.T) {
	database := newMigrationTestDatabase(t)

	all := []Migration{
		{
			Id: "20260101_01_half_done",
			Up: func(c core.Context, tx *MigrationTx) error {
				// The change succeeds, then the migration fails: the change must not
				// survive, otherwise the next run starts from a half-applied schema.
				if err := tx.Exec("CREATE TABLE half_done (id INTEGER PRIMARY KEY)"); err != nil {
					return err
				}

				return assert.AnError
			},
		},
	}

	appliedNow, err := runMigrations(core.NewNullContext(), database, all)

	assert.NotNil(t, err)
	assert.Empty(t, appliedNow)

	hasTable, err := database.HasTable("half_done")
	require.Nil(t, err)
	assert.False(t, hasTable, "the table created by the failed migration must be rolled back")

	var records []*models.SchemaMigration

	session := database.NewSession(core.NewNullContext())
	defer session.Close()

	require.Nil(t, session.Find(&records))
	assert.Empty(t, records)

	// Retrying with a fixed migration of the same id is enough to recover.
	all[0].Up = func(c core.Context, tx *MigrationTx) error {
		return tx.Exec("CREATE TABLE half_done (id INTEGER PRIMARY KEY)")
	}

	appliedNow, err = runMigrations(core.NewNullContext(), database, all)
	require.Nil(t, err)
	assert.Equal(t, []string{"20260101_01_half_done"}, appliedNow)

	hasTable, err = database.HasTable("half_done")
	require.Nil(t, err)
	assert.True(t, hasTable)
}

func TestMigrationTxSeesSchemaInsideTheTransaction(t *testing.T) {
	database := newMigrationTestDatabase(t)

	require.Nil(t, database.Exec(core.NewNullContext(), "CREATE TABLE existing (id INTEGER PRIMARY KEY, name VARCHAR(32))"))

	var sawTable, sawCreatedTable bool

	all := []Migration{
		{
			Id: "20260101_01_inspect",
			Up: func(c core.Context, tx *MigrationTx) error {
				hasTable, err := tx.HasTable("existing")

				if err != nil {
					return err
				}

				hasColumn, err := tx.HasColumn("existing", "name")

				if err != nil {
					return err
				}

				sawTable = hasTable && hasColumn

				// A change made earlier in the same migration is visible too, which
				// is what lets a migration branch on what it has already done.
				if err := tx.Exec("CREATE TABLE created_here (id INTEGER PRIMARY KEY)"); err != nil {
					return err
				}

				created, err := tx.HasTable("created_here")

				if err != nil {
					return err
				}

				sawCreatedTable = created

				return nil
			},
		},
	}

	_, err := runMigrations(core.NewNullContext(), database, all)
	require.Nil(t, err)
	assert.True(t, sawTable)
	assert.True(t, sawCreatedTable)
}

func TestHasColumnHelper(t *testing.T) {
	database := newMigrationTestDatabase(t)

	require.Nil(t, database.Exec(core.NewNullContext(), "CREATE TABLE widget (id INTEGER PRIMARY KEY, name VARCHAR(32))"))

	hasTable, err := database.HasTable("widget")
	require.Nil(t, err)
	assert.True(t, hasTable)

	hasColumn, err := database.HasColumn("widget", "name")
	require.Nil(t, err)
	assert.True(t, hasColumn)

	missingColumn, err := database.HasColumn("widget", "colour")
	require.Nil(t, err)
	assert.False(t, missingColumn)

	missingTable, err := database.HasTable("not_here")
	require.Nil(t, err)
	assert.False(t, missingTable)
}
