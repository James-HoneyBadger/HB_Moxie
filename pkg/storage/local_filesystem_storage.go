package storage

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/settings"
	"github.com/James-HoneyBadger/honey-ledger/pkg/utils"
)

// LocalFileSystemObjectStorage represents local file system object storage
type LocalFileSystemObjectStorage struct {
	rootPath string
}

// NewLocalFileSystemObjectStorage returns a local file system object storage
func NewLocalFileSystemObjectStorage(config *settings.Config, pathPrefix string) (*LocalFileSystemObjectStorage, error) {
	if err := os.MkdirAll(config.LocalFileSystemPath, 0750); err != nil {
		return nil, err
	}

	if err := os.Chmod(config.LocalFileSystemPath, 0750); err != nil {
		return nil, err
	}

	storage := &LocalFileSystemObjectStorage{
		rootPath: filepath.Join(config.LocalFileSystemPath, pathPrefix),
	}

	if err := os.MkdirAll(storage.rootPath, 0750); err != nil {
		return nil, err
	}

	if err := os.Chmod(storage.rootPath, 0750); err != nil {
		return nil, err
	}

	return storage, nil
}

// Exists returns whether the file exists
func (s *LocalFileSystemObjectStorage) Exists(ctx core.Context, path string) (bool, error) {
	return utils.IsExists(s.getFinalPath(path))
}

// Read returns the object instance according to specified the file path
func (s *LocalFileSystemObjectStorage) Read(ctx core.Context, path string) (ObjectInStorage, error) {
	return os.Open(s.getFinalPath(path))
}

// Save returns whether save the object instance successfully
func (s *LocalFileSystemObjectStorage) Save(ctx core.Context, path string, object ObjectInStorage) error {
	finalPath := s.getFinalPath(path)

	targetDir := filepath.Dir(finalPath)

	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return err
	}

	if err := os.Chmod(targetDir, 0750); err != nil {
		return err
	}

	targetFile, err := os.OpenFile(finalPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)

	if err != nil {
		return err
	}

	defer targetFile.Close()

	if err := targetFile.Chmod(0600); err != nil {
		return err
	}

	_, err = io.Copy(targetFile, object)

	return err
}

// Delete returns whether delete the object according to specified the file path successfully
func (s *LocalFileSystemObjectStorage) Delete(ctx core.Context, path string) error {
	return os.Remove(s.getFinalPath(path))
}

func (s *LocalFileSystemObjectStorage) getFinalPath(path string) string {
	finalPath := filepath.Join(s.rootPath, path)
	cleanPath := filepath.Clean(finalPath)

	if !strings.HasPrefix(cleanPath, filepath.Clean(s.rootPath)+string(filepath.Separator)) && cleanPath != filepath.Clean(s.rootPath) {
		return s.rootPath
	}

	return cleanPath
}
