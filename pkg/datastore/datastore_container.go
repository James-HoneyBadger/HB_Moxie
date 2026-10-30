package datastore

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"

	"github.com/James-HoneyBadger/honey-ledger/pkg/errs"
	"github.com/James-HoneyBadger/honey-ledger/pkg/settings"
)

// DataStoreContainer contains all data storages
type DataStoreContainer struct {
	UserStore     *DataStore
	TokenStore    *DataStore
	UserDataStore *DataStore
}

// Initialize a data storage container singleton instance
var (
	Container = &DataStoreContainer{}
)

// InitializeDataStore initializes data storage according to the config
func InitializeDataStore(config *settings.Config) error {
	database, err := initializeDatabase(config.DatabaseConfig)

	if err != nil {
		return err
	}

	setDatabaseLogger(database, config)

	Container.UserStore, err = NewDataStore(database)

	if err != nil {
		return err
	}

	Container.TokenStore, err = NewDataStore(database)

	if err != nil {
		return err
	}

	Container.UserDataStore, err = NewDataStore(database)

	if err != nil {
		return err
	}

	return nil
}

func initializeDatabase(dbConfig *settings.DatabaseConfig) (*Database, error) {
	var connStr string
	var err error

	if dbConfig.DatabaseType == settings.Sqlite3DbType {
		databaseDir := filepath.Dir(dbConfig.DatabasePath)

		if err = os.MkdirAll(databaseDir, 0750); err != nil {
			return nil, err
		}

		if err = os.Chmod(databaseDir, 0750); err != nil {
			return nil, err
		}

		if _, err = os.Stat(dbConfig.DatabasePath); err != nil {
			file, err := os.OpenFile(dbConfig.DatabasePath, os.O_CREATE|os.O_WRONLY, 0600)

			if err != nil {
				return nil, err
			}

			defer file.Close()
		}

		if err = os.Chmod(dbConfig.DatabasePath, 0600); err != nil {
			return nil, err
		}

		// Tighten pre-existing WAL/SHM sidecars as well; new ones inherit the
		// restrictive process umask.
		for _, suffix := range []string{"-wal", "-shm"} {
			_ = os.Chmod(dbConfig.DatabasePath+suffix, 0600)
		}
	}

	if dbConfig.DatabaseType == settings.MySqlDbType {
		connStr, err = getMysqlConnectionString(dbConfig)
	} else if dbConfig.DatabaseType == settings.PostgresDbType {
		connStr, err = getPostgresConnectionString(dbConfig)
	} else if dbConfig.DatabaseType == settings.Sqlite3DbType {
		connStr, err = getSqlite3ConnectionString(dbConfig)
	} else {
		return nil, errs.ErrDatabaseTypeInvalid
	}

	if err != nil {
		return nil, err
	}

	connStrs := []string{
		connStr,
	}
	engineGroup, err := xorm.NewEngineGroup(dbConfig.DatabaseType, connStrs, xorm.RoundRobinPolicy())

	if err != nil {
		return nil, err
	}

	maxIdleConns := int(dbConfig.MaxIdleConnection)
	maxOpenConns := int(dbConfig.MaxOpenConnection)

	if dbConfig.DatabaseType == settings.Sqlite3DbType && maxOpenConns == 0 {
		// SQLite serializes writers. An unbounded pool turns concurrent writes
		// into avoidable SQLITE_BUSY errors, so keep a single connection.
		maxOpenConns = 1
	}

	if maxOpenConns > 0 && maxIdleConns > maxOpenConns {
		maxIdleConns = maxOpenConns
	}

	engineGroup.SetMaxIdleConns(maxIdleConns)
	engineGroup.SetMaxOpenConns(maxOpenConns)
	engineGroup.SetConnMaxLifetime(time.Duration(dbConfig.ConnectionMaxLifeTime) * time.Second)

	return &Database{
		databaseType: dbConfig.DatabaseType,
		engineGroup:  engineGroup,
	}, nil
}

func setDatabaseLogger(database *Database, config *settings.Config) {
	if config.EnableQueryLog {
		database.engineGroup.SetLogger(NewXOrmLoggerAdapter(config.EnableQueryLog, config.LogLevel))
		database.engineGroup.ShowSQL(true)
	}
}

func getMysqlConnectionString(dbConfig *settings.DatabaseConfig) (string, error) {
	protocol := "tcp"

	if strings.HasPrefix(dbConfig.DatabaseHost, "/") { // unix socket path
		protocol = "unix"
	}

	return fmt.Sprintf("%s:%s@%s(%s)/%s?charset=utf8mb4&parseTime=true",
		dbConfig.DatabaseUser, dbConfig.DatabasePassword, protocol, dbConfig.DatabaseHost, dbConfig.DatabaseName), nil
}

func getPostgresConnectionString(dbConfig *settings.DatabaseConfig) (string, error) {
	if strings.HasPrefix(dbConfig.DatabaseHost, "/") { // unix socket path
		return fmt.Sprintf("postgres:///%s?sslmode=%s&host=%s&user=%s&password=%s",
			dbConfig.DatabaseName, dbConfig.DatabaseSSLMode, dbConfig.DatabaseHost, url.QueryEscape(dbConfig.DatabaseUser), url.QueryEscape(dbConfig.DatabasePassword)), nil
	} else {
		host, port, err := net.SplitHostPort(dbConfig.DatabaseHost)

		if err != nil {
			return "", errs.ErrDatabaseHostInvalid
		}

		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			url.QueryEscape(dbConfig.DatabaseUser), url.QueryEscape(dbConfig.DatabasePassword), host, port, dbConfig.DatabaseName, dbConfig.DatabaseSSLMode), nil
	}
}

func getSqlite3ConnectionString(dbConfig *settings.DatabaseConfig) (string, error) {
	return fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&cache=shared&mode=rwc", dbConfig.DatabasePath), nil
}

// Close closes all database connections held by the container
func (c *DataStoreContainer) Close() error {
	seen := make(map[*Database]bool)

	for _, store := range []*DataStore{c.UserStore, c.TokenStore, c.UserDataStore} {
		if store == nil {
			continue
		}

		for i := 0; i < store.Count(); i++ {
			database := store.Get(i)

			if database == nil || seen[database] {
				continue
			}

			seen[database] = true

			if err := database.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}
