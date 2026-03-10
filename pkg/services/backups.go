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

	"github.com/James-HoneyBadger/HB_Moxie/pkg/core"
	"github.com/James-HoneyBadger/HB_Moxie/pkg/errs"
	"github.com/James-HoneyBadger/HB_Moxie/pkg/models"
	"github.com/James-HoneyBadger/HB_Moxie/pkg/settings"
	"github.com/James-HoneyBadger/HB_Moxie/pkg/utils"
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
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), backupFileExtension) {
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
	backupDir, err := s.GetBackupDir()

	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Format(backupFileNameTimeFormat)
	fileName := fmt.Sprintf("backup_%s%s", timestamp, backupFileExtension)
	backupPath := filepath.Join(backupDir, fileName)

	zipFile, err := os.Create(backupPath)

	if err != nil {
		return nil, err
	}

	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Backup SQLite database if applicable
	if config.DatabaseConfig != nil && config.DatabaseConfig.DatabaseType == settings.Sqlite3DbType && config.DatabaseConfig.DatabasePath != "" {
		dbPath := config.DatabaseConfig.DatabasePath

		if err := addFileToZip(zipWriter, dbPath, "database/"+filepath.Base(dbPath)); err != nil {
			os.Remove(backupPath)
			return nil, fmt.Errorf("failed to backup database: %w", err)
		}
	}

	// Backup storage directory
	storagePath := config.LocalFileSystemPath

	if storagePath != "" {
		if err := addDirectoryToZip(zipWriter, storagePath, "storage"); err != nil {
			os.Remove(backupPath)
			return nil, fmt.Errorf("failed to backup storage: %w", err)
		}
	}

	// Backup configuration
	confPath := filepath.Join(config.WorkingPath, "conf", "hbmoxie.ini")

	if _, statErr := os.Stat(confPath); statErr == nil {
		if err := addFileToZip(zipWriter, confPath, "conf/hbmoxie.ini"); err != nil {
			os.Remove(backupPath)
			return nil, fmt.Errorf("failed to backup config: %w", err)
		}
	}

	// Close the zip writer to flush
	zipWriter.Close()
	zipFile.Close()

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

// GetBackupFileContent reads and returns the backup file content
func (s *BackupService) GetBackupFileContent(c core.Context, fileName string) ([]byte, error) {
	if !isValidBackupFileName(fileName) {
		return nil, errs.ErrBackupInvalidFileName
	}

	backupDir, err := s.GetBackupDir()

	if err != nil {
		return nil, err
	}

	backupPath := filepath.Join(backupDir, fileName)

	// Security: prevent path traversal
	absBackupDir, _ := filepath.Abs(backupDir)
	absBackupPath, _ := filepath.Abs(backupPath)

	if !strings.HasPrefix(absBackupPath, absBackupDir) {
		return nil, errs.ErrBackupInvalidFileName
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil, errs.ErrBackupNotFound
	}

	return os.ReadFile(backupPath)
}

// RestoreBackup restores from a backup archive
func (s *BackupService) RestoreBackup(c core.Context, fileName string) error {
	if !isValidBackupFileName(fileName) {
		return errs.ErrBackupInvalidFileName
	}

	config := s.CurrentConfig()
	backupDir, err := s.GetBackupDir()

	if err != nil {
		return err
	}

	backupPath := filepath.Join(backupDir, fileName)

	// Security: prevent path traversal
	absBackupDir, _ := filepath.Abs(backupDir)
	absBackupPath, _ := filepath.Abs(backupPath)

	if !strings.HasPrefix(absBackupPath, absBackupDir) {
		return errs.ErrBackupInvalidFileName
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

	// Restore database
	if config.DatabaseConfig != nil && config.DatabaseConfig.DatabaseType == settings.Sqlite3DbType && config.DatabaseConfig.DatabasePath != "" {
		dbBaseName := filepath.Base(config.DatabaseConfig.DatabasePath)

		for _, f := range reader.File {
			if f.Name == "database/"+dbBaseName {
				if err := restoreFileFromZip(f, config.DatabaseConfig.DatabasePath); err != nil {
					return fmt.Errorf("failed to restore database: %w", err)
				}

				break
			}
		}
	}

	// Restore storage files
	if config.LocalFileSystemPath != "" {
		for _, f := range reader.File {
			if strings.HasPrefix(f.Name, "storage/") && !f.FileInfo().IsDir() {
				relativePath := strings.TrimPrefix(f.Name, "storage/")
				destPath := filepath.Join(config.LocalFileSystemPath, relativePath)

				// Security: ensure destination is within storage path
				absStorage, _ := filepath.Abs(config.LocalFileSystemPath)
				absDest, _ := filepath.Abs(destPath)

				if !strings.HasPrefix(absDest, absStorage) {
					continue
				}

				if err := restoreFileFromZip(f, destPath); err != nil {
					return fmt.Errorf("failed to restore storage file %s: %w", relativePath, err)
				}
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

	backupPath := filepath.Join(backupDir, fileName)

	// Security: prevent path traversal
	absBackupDir, _ := filepath.Abs(backupDir)
	absBackupPath, _ := filepath.Abs(backupPath)

	if !strings.HasPrefix(absBackupPath, absBackupDir) {
		return errs.ErrBackupInvalidFileName
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return errs.ErrBackupNotFound
	}

	return os.Remove(backupPath)
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

	destFile, err := os.Create(destPath)

	if err != nil {
		return err
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, rc)
	return err
}

// Unused import guard
var _ = utils.Int64ToString
