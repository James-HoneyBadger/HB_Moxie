package datastore

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"

	"github.com/James-HoneyBadger/ratel-ledger/pkg/core"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/settings"
)

func TestGetMysqlConnectionString_TCP(t *testing.T) {
	expectedValue := "username:password@tcp(1.2.3.4:3306)/dbname?charset=utf8mb4&parseTime=true"
	actualValue, err := getMysqlConnectionString(&settings.DatabaseConfig{
		DatabaseType:     "mysql",
		DatabaseHost:     "1.2.3.4:3306",
		DatabaseName:     "dbname",
		DatabaseUser:     "username",
		DatabasePassword: "password",
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetMysqlConnectionString_UnixSocket(t *testing.T) {
	expectedValue := "username:password@unix(/path/to/mysql.sock)/dbname?charset=utf8mb4&parseTime=true"
	actualValue, err := getMysqlConnectionString(&settings.DatabaseConfig{
		DatabaseType:     "mysql",
		DatabaseHost:     "/path/to/mysql.sock",
		DatabaseName:     "dbname",
		DatabaseUser:     "username",
		DatabasePassword: "password",
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetPostgreSQLConnectionString_TCP(t *testing.T) {
	expectedValue := "postgres://username:password@1.2.3.4:5432/dbname?sslmode=disable"
	actualValue, err := getPostgresConnectionString(&settings.DatabaseConfig{
		DatabaseType:     "postgres",
		DatabaseHost:     "1.2.3.4:5432",
		DatabaseName:     "dbname",
		DatabaseUser:     "username",
		DatabasePassword: "password",
		DatabaseSSLMode:  "disable",
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetPostgreSQLConnectionString_UnixSocket(t *testing.T) {
	expectedValue := "postgres:///dbname?sslmode=disable&host=/path/to/postgres.sock&user=username&password=password"
	actualValue, err := getPostgresConnectionString(&settings.DatabaseConfig{
		DatabaseType:     "postgres",
		DatabaseHost:     "/path/to/postgres.sock",
		DatabaseName:     "dbname",
		DatabaseUser:     "username",
		DatabasePassword: "password",
		DatabaseSSLMode:  "disable",
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedValue, actualValue)
}

func TestGetSqlite3ConnectionString_EnablesReliableConcurrencySettings(t *testing.T) {
	actualValue, err := getSqlite3ConnectionString(&settings.DatabaseConfig{
		DatabaseType: settings.Sqlite3DbType,
		DatabasePath: "/tmp/ratel-ledger.db",
	})

	assert.Nil(t, err)
	assert.Contains(t, actualValue, "_journal_mode=WAL")
	assert.Contains(t, actualValue, "_busy_timeout=5000")
	assert.Contains(t, actualValue, "_synchronous=NORMAL")
}

func TestVacuumInto_CreatesConsistentSnapshot(t *testing.T) {
	workingDir := t.TempDir()
	sourcePath := filepath.Join(workingDir, "source.db")
	snapshotPath := filepath.Join(workingDir, "snapshot.db")

	dsn, err := getSqlite3ConnectionString(&settings.DatabaseConfig{
		DatabaseType: settings.Sqlite3DbType,
		DatabasePath: sourcePath,
	})
	assert.Nil(t, err)

	engineGroup, err := xorm.NewEngineGroup(settings.Sqlite3DbType, []string{dsn}, xorm.RoundRobinPolicy())
	assert.Nil(t, err)
	defer engineGroup.Close()

	_, err = engineGroup.Exec("CREATE TABLE sample (id INTEGER PRIMARY KEY, name TEXT NOT NULL)")
	assert.Nil(t, err)

	_, err = engineGroup.Exec("INSERT INTO sample (id, name) VALUES (1, 'value')")
	assert.Nil(t, err)

	database := &Database{
		databaseType: settings.Sqlite3DbType,
		engineGroup:  engineGroup,
	}

	err = database.VacuumInto(core.NewNullContext(), snapshotPath)
	assert.Nil(t, err)

	snapshotEngineGroup, err := xorm.NewEngineGroup(settings.Sqlite3DbType, []string{fmt.Sprintf("file:%s?mode=ro", snapshotPath)}, xorm.RoundRobinPolicy())
	assert.Nil(t, err)
	defer snapshotEngineGroup.Close()

	rows, err := snapshotEngineGroup.QueryString("SELECT name FROM sample WHERE id = ?", 1)
	assert.Nil(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, "value", rows[0]["name"])
}
