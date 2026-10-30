package cmd

import (
	"github.com/urfave/cli/v3"

	"github.com/James-HoneyBadger/ratel-ledger/pkg/core"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/datastore"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/log"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/models"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/settings"
)

// Database represents the database command
var Database = &cli.Command{
	Name:  "database",
	Usage: "Ratel Ledger database maintenance",
	Commands: []*cli.Command{
		{
			Name:   "update",
			Usage:  "Update database structure",
			Action: bindAction(updateDatabaseStructure),
		},
	},
}

func updateDatabaseStructure(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	log.CliInfof(c, "[database.updateDatabaseStructure] starting maintaining")

	err = updateAllDatabaseTablesStructure(c)

	if err != nil {
		log.CliErrorf(c, "[database.updateDatabaseStructure] update database table structure failed, because %s", err.Error())
		return err
	}

	log.CliInfof(c, "[database.updateDatabaseStructure] all tables maintained successfully")
	return nil
}

func updateAllDatabaseTablesStructure(c *core.CliContext) error {
	var err error

	err = datastore.Container.UserStore.SyncStructs(new(models.User))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user table maintained successfully")

	err = ensureUserCredentialColumnLengths(c)

	if err != nil {
		log.BootErrorf(c, "[database.updateAllDatabaseTablesStructure] user credential column migration failed, because %s", err.Error())
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user credential columns checked successfully")

	err = datastore.Container.UserStore.SyncStructs(new(models.TwoFactor))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] two-factor table maintained successfully")

	err = datastore.Container.UserStore.SyncStructs(new(models.TwoFactorRecoveryCode))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] two-factor recovery code table maintained successfully")

	err = datastore.Container.TokenStore.SyncStructs(new(models.TokenRecord))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] token record table maintained successfully")

	err = ensureTokenSecretColumnLength(c)

	if err != nil {
		log.BootErrorf(c, "[database.updateAllDatabaseTablesStructure] token record secret column migration failed, because %s", err.Error())
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] token record secret column checked successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.Account))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] account table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.Transaction))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionCategory))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction category table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionTagGroup))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction tag group table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionTag))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction tag table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionTagIndex))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction tag index table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionTemplate))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction template table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionPictureInfo))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction picture table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.UserCustomExchangeRate))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user custom exchange rate table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.UserApplicationCloudSetting))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user application cloud settings table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.UserExternalAuth))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user external auth table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.InsightsExplorer))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] insights explorer table maintained successfully")

	appliedMigrations, err := datastore.Container.RunMigrations(c)

	if err != nil {
		log.BootErrorf(c, "[database.updateAllDatabaseTablesStructure] database migrations failed, because %s", err.Error())
		return err
	}

	for _, migrationId := range appliedMigrations {
		log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] applied database migration \"%s\"", migrationId)
	}

	return nil
}

// ensureUserCredentialColumnLengths widens the password and salt columns for
// MySQL and PostgreSQL databases created before versioned password hashes.
func ensureUserCredentialColumnLengths(c *core.CliContext) error {
	config := settings.Container.GetCurrentConfig()

	if config == nil || config.DatabaseConfig == nil {
		return nil
	}

	var statements []string

	switch config.DatabaseConfig.DatabaseType {
	case settings.MySqlDbType:
		statements = []string{
			"ALTER TABLE `user` MODIFY COLUMN password VARCHAR(128) NOT NULL",
			"ALTER TABLE `user` MODIFY COLUMN salt VARCHAR(32) NOT NULL",
		}
	case settings.PostgresDbType:
		statements = []string{
			"ALTER TABLE \"user\" ALTER COLUMN password TYPE VARCHAR(128)",
			"ALTER TABLE \"user\" ALTER COLUMN salt TYPE VARCHAR(32)",
		}
	default:
		// SQLite does not enforce VARCHAR lengths.
		return nil
	}

	for _, statement := range statements {
		if err := datastore.Container.UserStore.Choose(0).Exec(c, statement); err != nil {
			return err
		}
	}

	return nil
}

// ensureTokenSecretColumnLength widens token_record.secret for MySQL and PostgreSQL
// databases created before token secrets were increased to 32 characters.
func ensureTokenSecretColumnLength(c *core.CliContext) error {
	config := settings.Container.GetCurrentConfig()

	if config == nil || config.DatabaseConfig == nil {
		return nil
	}

	var statement string

	switch config.DatabaseConfig.DatabaseType {
	case settings.MySqlDbType:
		statement = "ALTER TABLE token_record MODIFY COLUMN secret VARCHAR(64) NOT NULL"
	case settings.PostgresDbType:
		statement = "ALTER TABLE token_record ALTER COLUMN secret TYPE VARCHAR(64)"
	default:
		// SQLite does not enforce VARCHAR lengths.
		return nil
	}

	return datastore.Container.TokenStore.Choose(0).Exec(c, statement)
}
