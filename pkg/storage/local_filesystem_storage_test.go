package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/James-HoneyBadger/ratel-ledger/pkg/settings"
)

type testStorageObject struct {
	*strings.Reader
}

func (testStorageObject) Close() error {
	return nil
}

func TestLocalFileSystemObjectStorage_PrivatePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not enforced on Windows")
	}

	rootPath := t.TempDir()
	objectStorage, err := NewLocalFileSystemObjectStorage(&settings.Config{LocalFileSystemPath: rootPath}, "pictures")
	assert.NoError(t, err)

	err = objectStorage.Save(nil, "user/avatar.png", testStorageObject{strings.NewReader("image-data")})
	assert.NoError(t, err)

	storageRootInfo, err := os.Stat(rootPath)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0750), storageRootInfo.Mode().Perm())

	rootInfo, err := os.Stat(filepath.Join(rootPath, "pictures"))
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0750), rootInfo.Mode().Perm())

	dirInfo, err := os.Stat(filepath.Join(rootPath, "pictures", "user"))
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0750), dirInfo.Mode().Perm())

	fileInfo, err := os.Stat(filepath.Join(rootPath, "pictures", "user", "avatar.png"))
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), fileInfo.Mode().Perm())
}
