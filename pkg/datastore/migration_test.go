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
			Up: func(c core.Context, db *Database) error {
				applied = append(applied, "second")
				return db.Exec(c, "CREATE TABLE second_table (id INTEGER PRIMARY KEY)")
			},
		},
		{
			Id:          "20260101_01_first",
			Description: "first",
			Up: func(c core.Context, db *Database) error {
				applied = append(applied, "first")
				return db.Exec(c, "CREATE TABLE first_table (id INTEGER PRIMARY KEY)")
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
			Up: func(c core.Context, db *Database) error {
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
