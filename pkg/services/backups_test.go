package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/datastore"
	"github.com/James-HoneyBadger/honey-ledger/pkg/settings"
)

func TestIsPathWithinDirectory(t *testing.T) {
	baseDir := filepath.Join(string(filepath.Separator), "data", "storage")

	assert.True(t, isPathWithinDirectory(baseDir, filepath.Join(baseDir, "avatars", "1.png")))
	assert.False(t, isPathWithinDirectory(baseDir, filepath.Join(string(filepath.Separator), "data", "storage2", "1.png")))
	assert.False(t, isPathWithinDirectory(baseDir, filepath.Join(baseDir, "..", "secret")))
	assert.False(t, isPathWithinDirectory(baseDir, baseDir))
}

func TestResolveBackupPath_RejectsTraversal(t *testing.T) {
	_, err := resolveBackupPath(filepath.Join(string(filepath.Separator), "data", "backups"), ".."+string(filepath.Separator)+"backup_20240101_000000.zip")

	assert.Error(t, err)
}

func TestResolveBackupPath_AcceptsValidName(t *testing.T) {
	backupDir := filepath.Join(string(filepath.Separator), "data", "backups")
	actual, err := resolveBackupPath(backupDir, "backup_20240101_000000.zip")

	assert.Nil(t, err)
	assert.Equal(t, filepath.Join(backupDir, "backup_20240101_000000.zip"), actual)
}

func TestCreateAndRestoreBackup_SQLite(t *testing.T) {
	workingDir := t.TempDir()
	databaseDir := filepath.Join(workingDir, "data")
	storageDir := filepath.Join(workingDir, "storage")

	assert.Nil(t, os.MkdirAll(databaseDir, 0750))
	assert.Nil(t, os.MkdirAll(storageDir, 0750))

	config := &settings.Config{
		WorkingPath: workingDir,
		DatabaseConfig: &settings.DatabaseConfig{
			DatabaseType:      settings.Sqlite3DbType,
			DatabasePath:      filepath.Join(databaseDir, "honey-ledger.db"),
			MaxIdleConnection: 1,
			MaxOpenConnection: 1,
		},
		LocalFileSystemPath: storageDir,
	}

	previousConfig := settings.Container.GetCurrentConfig()
	settings.SetCurrentConfig(config)
	assert.Nil(t, datastore.InitializeDataStore(config))

	defer func() {
		_ = datastore.Container.Close()
		settings.SetCurrentConfig(previousConfig)
	}()

	ctx := core.NewNullContext()
	database := datastore.Container.UserDataStore.Choose(0)

	assert.Nil(t, database.Exec(ctx, "CREATE TABLE backup_test (id INTEGER PRIMARY KEY, value TEXT NOT NULL)"))
	assert.Nil(t, database.Exec(ctx, "INSERT INTO backup_test (id, value) VALUES (1, 'before')"))

	storageFilePath := filepath.Join(storageDir, "sample.txt")
	assert.Nil(t, os.WriteFile(storageFilePath, []byte("payload"), 0600))

	backup, err := Backups.CreateBackup(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, backup)

	assert.Nil(t, database.Exec(ctx, "UPDATE backup_test SET value = 'after' WHERE id = 1"))
	assert.Nil(t, os.WriteFile(storageFilePath, []byte("changed"), 0600))

	assert.Nil(t, Backups.RestoreBackup(ctx, backup.FileName))

	rows, err := datastore.Container.UserDataStore.Choose(0).NewSession(ctx).QueryString("SELECT value FROM backup_test WHERE id = ?", 1)
	assert.Nil(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, "before", rows[0]["value"])

	restoredStorageContent, err := os.ReadFile(storageFilePath)
	assert.Nil(t, err)
	assert.Equal(t, "payload", string(restoredStorageContent))
}
