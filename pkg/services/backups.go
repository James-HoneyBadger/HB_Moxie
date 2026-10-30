package services

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/datastore"
	"github.com/James-HoneyBadger/honey-ledger/pkg/errs"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/settings"
)

const backupDirName = "backups"
const backupFileExtension = ".zip"
const backupFileNameTimeFormat = "20060102_150405"

var validBackupFileNamePattern = regexp.MustCompile(`^backup_\d{8}_\d{6}\.zip$`)

// BackupService represents backup service
type BackupService struct {
	ServiceUsingConfig
}

// Backups is the backup service singleton
var Backups = &BackupService{
	ServiceUsingConfig: ServiceUsingConfig{
		container: settings.Container,
	},
}

// GetBackupDir returns the absolute path of the backup directory
func (s *BackupService) GetBackupDir() (string, error) {
	config := s.CurrentConfig()
	backupDir := filepath.Join(config.WorkingPath, backupDirName)

	if err := os.MkdirAll(backupDir, 0750); err != nil {
		return "", err
	}

	if err := os.Chmod(backupDir, 0750); err != nil {
		return "", err
	}

	return backupDir, nil
}

// ListBackups returns all backup files
func (s *BackupService) ListBackups(c core.Context) ([]*models.BackupListResponse, error) {
	backupDir, err := s.GetBackupDir()

	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(backupDir)

	if err != nil {
		return nil, err
	}

	var backups []*models.BackupListResponse

	for _, entry := range entries {
		if entry.IsDir() || !isValidBackupFileName(entry.Name()) {
			continue
		}

		info, err := entry.Info()

		if err != nil {
			continue
		}

		backups = append(backups, &models.BackupListResponse{
			FileName:  entry.Name(),
			FileSize:  info.Size(),
			CreatedAt: info.ModTime().Unix(),
		})
	}

	// Sort by newest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt > backups[j].CreatedAt
	})

	return backups, nil
}

// CreateBackup creates a new backup archive containing the database and storage files
func (s *BackupService) CreateBackup(c core.Context) (*models.BackupListResponse, error) {
	config := s.CurrentConfig()

	if config.DatabaseConfig == nil || config.DatabaseConfig.DatabaseType != settings.Sqlite3DbType || config.DatabaseConfig.DatabasePath == "" {
		return nil, errs.ErrBackupDatabaseTypeNotSupported
	}

	backupDir, err := s.GetBackupDir()

	if err != nil {
		return nil, err
	}

	fileName, backupPath, zipFile, err := createBackupFile(backupDir)

	if err != nil {
		return nil, err
	}

	zipWriter := zip.NewWriter(zipFile)
	completed := false

	defer func() {
		if !completed {
			_ = zipWriter.Close()
			_ = zipFile.Close()
			_ = os.Remove(backupPath)
		}
	}()

	// Use SQLite's online backup mechanism instead of copying a live database file.
	snapshotPath, err := s.createDatabaseSnapshot(c, backupDir)

	if err != nil {
		return nil, fmt.Errorf("failed to snapshot database: %w", err)
	}

	defer os.Remove(snapshotPath)

	if err := addFileToZip(zipWriter, snapshotPath, "database/"+filepath.Base(config.DatabaseConfig.DatabasePath)); err != nil {
		return nil, fmt.Errorf("failed to backup database: %w", err)
	}

	if config.LocalFileSystemPath != "" {
		info, statErr := os.Stat(config.LocalFileSystemPath)

		if statErr == nil && info.IsDir() {
			if err := addDirectoryToZip(zipWriter, config.LocalFileSystemPath, "storage"); err != nil {
				return nil, fmt.Errorf("failed to backup storage: %w", err)
			}
		} else if statErr != nil && !os.IsNotExist(statErr) {
			return nil, fmt.Errorf("failed to access storage directory: %w", statErr)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	if err := zipFile.Close(); err != nil {
		return nil, err
	}

	completed = true

	info, err := os.Stat(backupPath)

	if err != nil {
		return nil, err
	}

	return &models.BackupListResponse{
		FileName:  fileName,
		FileSize:  info.Size(),
		CreatedAt: info.ModTime().Unix(),
	}, nil
}

func (s *BackupService) createDatabaseSnapshot(c core.Context, backupDir string) (string, error) {
	snapshotPath := filepath.Join(backupDir, fmt.Sprintf(".snapshot-%d.db", time.Now().UnixNano()))

	if err := os.Remove(snapshotPath); err != nil && !os.IsNotExist(err) {
		return "", err
	}

	if err := datastore.Container.UserDataStore.Choose(0).VacuumInto(c, snapshotPath); err != nil {
		return "", err
	}

	return snapshotPath, nil
}

func createBackupFile(backupDir string) (string, string, *os.File, error) {
	for attempt := 0; attempt < 5; attempt++ {
		timestamp := time.Now().Add(time.Duration(attempt) * time.Second).Format(backupFileNameTimeFormat)
		fileName := fmt.Sprintf("backup_%s%s", timestamp, backupFileExtension)
		backupPath := filepath.Join(backupDir, fileName)
		file, err := os.OpenFile(backupPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)

		if err == nil {
			return fileName, backupPath, file, nil
		}

		if !os.IsExist(err) {
			return "", "", nil, err
		}
	}

	return "", "", nil, fmt.Errorf("failed to allocate a unique backup file name")
}

// GetBackupFileContent reads and returns the backup file content
func (s *BackupService) GetBackupFileContent(c core.Context, fileName string) ([]byte, error) {
	if !isValidBackupFileName(fileName) {
		return nil, errs.ErrBackupInvalidFileName
	}

	backupDir, err := s.GetBackupDir()

	if err != nil {
		return nil, err
	}

	backupPath, err := resolveBackupPath(backupDir, fileName)

	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil, errs.ErrBackupNotFound
	}

	return os.ReadFile(backupPath)
}

// RestoreBackup restores from a backup archive
func (s *BackupService) RestoreBackup(c core.Context, fileName string) (err error) {
	config := s.CurrentConfig()

	if config.DatabaseConfig == nil || config.DatabaseConfig.DatabaseType != settings.Sqlite3DbType || config.DatabaseConfig.DatabasePath == "" {
		return errs.ErrBackupDatabaseTypeNotSupported
	}

	backupDir, err := s.GetBackupDir()

	if err != nil {
		return err
	}

	backupPath, err := resolveBackupPath(backupDir, fileName)

	if err != nil {
		return err
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return errs.ErrBackupNotFound
	}

	data, err := os.ReadFile(backupPath)

	if err != nil {
		return err
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))

	if err != nil {
		return fmt.Errorf("failed to read backup archive: %w", err)
	}

	dbBaseName := filepath.Base(config.DatabaseConfig.DatabasePath)
	var databaseFile *zip.File

	for _, f := range reader.File {
		if f.Name == "database/"+dbBaseName {
			databaseFile = f
			break
		}
	}

	if databaseFile == nil {
		return fmt.Errorf("backup archive does not contain the database")
	}

	// The SQLite file cannot be replaced while the connection pool is open.
	// Reopen it as soon as the restore finishes, including on failure paths.
	closeErr := datastore.Container.Close()

	defer func() {
		if reopenErr := datastore.InitializeDataStore(config); reopenErr != nil && err == nil {
			err = fmt.Errorf("failed to reopen database after restore: %w", reopenErr)
		}
	}()

	if closeErr != nil {
		return fmt.Errorf("failed to close database before restore: %w", closeErr)
	}

	if err := restoreFileFromZip(databaseFile, config.DatabaseConfig.DatabasePath); err != nil {
		return fmt.Errorf("failed to restore database: %w", err)
	}

	if config.LocalFileSystemPath != "" {
		for _, f := range reader.File {
			if !strings.HasPrefix(f.Name, "storage/") || f.FileInfo().IsDir() {
				continue
			}

			relativePath := strings.TrimPrefix(f.Name, "storage/")
			destPath := filepath.Join(config.LocalFileSystemPath, relativePath)

			if !isPathWithinDirectory(config.LocalFileSystemPath, destPath) {
				continue
			}

			if err := restoreFileFromZip(f, destPath); err != nil {
				return fmt.Errorf("failed to restore storage file %s: %w", relativePath, err)
			}
		}
	}

	return nil
}

// DeleteBackup deletes a backup file
func (s *BackupService) DeleteBackup(c core.Context, fileName string) error {
	if !isValidBackupFileName(fileName) {
		return errs.ErrBackupInvalidFileName
	}

	backupDir, err := s.GetBackupDir()

	if err != nil {
		return err
	}

	backupPath, err := resolveBackupPath(backupDir, fileName)

	if err != nil {
		return err
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return errs.ErrBackupNotFound
	}

	return os.Remove(backupPath)
}

func resolveBackupPath(backupDir string, fileName string) (string, error) {
	if !isValidBackupFileName(fileName) {
		return "", errs.ErrBackupInvalidFileName
	}

	backupPath := filepath.Join(backupDir, fileName)

	if !isPathWithinDirectory(backupDir, backupPath) {
		return "", errs.ErrBackupInvalidFileName
	}

	return backupPath, nil
}

func isPathWithinDirectory(baseDir string, targetPath string) bool {
	absBaseDir, err := filepath.Abs(baseDir)

	if err != nil {
		return false
	}

	absTargetPath, err := filepath.Abs(targetPath)

	if err != nil {
		return false
	}

	relativePath, err := filepath.Rel(absBaseDir, absTargetPath)

	if err != nil || relativePath == "." || relativePath == ".." {
		return false
	}

	return !strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) && !filepath.IsAbs(relativePath)
}

func isValidBackupFileName(fileName string) bool {
	return validBackupFileNamePattern.MatchString(fileName)
}

func addFileToZip(zipWriter *zip.Writer, filePath string, nameInZip string) error {
	file, err := os.Open(filePath)

	if err != nil {
		return err
	}

	defer file.Close()

	info, err := file.Stat()

	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)

	if err != nil {
		return err
	}

	header.Name = nameInZip
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)

	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func addDirectoryToZip(zipWriter *zip.Writer, dirPath string, prefixInZip string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)

		if err != nil {
			return err
		}

		nameInZip := filepath.Join(prefixInZip, relPath)

		// Normalize path separators for zip
		nameInZip = strings.ReplaceAll(nameInZip, string(os.PathSeparator), "/")

		return addFileToZip(zipWriter, path, nameInZip)
	})
}

func restoreFileFromZip(f *zip.File, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0750); err != nil {
		return err
	}

	rc, err := f.Open()

	if err != nil {
		return err
	}

	defer rc.Close()

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)

	if err != nil {
		return err
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, rc)
	return err
}
