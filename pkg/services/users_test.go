package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/James-HoneyBadger/ratel-ledger/pkg/core"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/datastore"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/models"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/settings"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/uuid"
)

func TestUserPasswordHashing_NewAndLegacyUpgrade(t *testing.T) {
	workingDir := t.TempDir()
	databaseDir := filepath.Join(workingDir, "data")

	assert.Nil(t, os.MkdirAll(databaseDir, 0750))

	config := &settings.Config{
		WorkingPath: workingDir,
		DatabaseConfig: &settings.DatabaseConfig{
			DatabaseType:      settings.Sqlite3DbType,
			DatabasePath:      filepath.Join(databaseDir, "ratel-ledger.db"),
			MaxIdleConnection: 1,
			MaxOpenConnection: 1,
		},
		UuidGeneratorType: settings.InternalUuidGeneratorType,
		UuidServerId:      1,
	}

	previousConfig := settings.Container.GetCurrentConfig()
	settings.SetCurrentConfig(config)
	assert.Nil(t, datastore.InitializeDataStore(config))
	assert.Nil(t, uuid.InitializeUuidGenerator(config))

	defer func() {
		_ = datastore.Container.Close()
		settings.SetCurrentConfig(previousConfig)
	}()

	assert.Nil(t, datastore.Container.UserStore.SyncStructs(new(models.User)))

	ctx := core.NewNullContext()
	user := &models.User{
		Username:             "alice",
		Email:                "alice@example.com",
		Nickname:             "Alice",
		Password:             "secret-password",
		DefaultCurrency:      "USD",
		FirstDayOfWeek:       core.WEEKDAY_SUNDAY,
		TransactionEditScope: models.TRANSACTION_EDIT_SCOPE_ALL,
	}

	assert.Nil(t, Users.CreateUser(ctx, user, false))
	assert.Contains(t, user.Password, "pbkdf2-sha256$600000$")

	loggedIn, _, err := Users.GetUserByUsernameOrEmailAndPassword(ctx, "alice", "secret-password")
	assert.Nil(t, err)
	assert.Equal(t, user.Uid, loggedIn.Uid)

	// Simulate a user created before the versioned hash migration.
	legacyHash := "QrpKShMygoe4Ym4ibnA7cNDzCcSonBkgFl69IrtnDmp3oROft3/Td/DNXjsweosa"
	assert.Nil(t, datastore.Container.UserStore.Choose(0).Exec(ctx, "UPDATE `user` SET password=?, salt=? WHERE uid=?", legacyHash, "salt", user.Uid))

	loggedIn, _, err = Users.GetUserByUsernameOrEmailAndPassword(ctx, "alice", "foobar")
	assert.Nil(t, err)
	assert.Equal(t, user.Uid, loggedIn.Uid)
	assert.Contains(t, loggedIn.Password, "pbkdf2-sha256$600000$")

	storedUser := &models.User{}
	has, err := datastore.Container.UserStore.Choose(0).NewSession(ctx).ID(user.Uid).Get(storedUser)
	assert.Nil(t, err)
	assert.True(t, has)
	assert.Contains(t, storedUser.Password, "pbkdf2-sha256$600000$")
	assert.Equal(t, "salt", storedUser.Salt)
}
