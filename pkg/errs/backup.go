package errs

import (
	"net/http"
)

// Error codes related to backup
var (
	ErrBackupNotAllowed               = NewNormalError(NormalSubcategoryDataManagement, 4, http.StatusBadRequest, "backup not allowed")
	ErrBackupNotFound                 = NewNormalError(NormalSubcategoryDataManagement, 5, http.StatusNotFound, "backup file not found")
	ErrBackupCreationFailed           = NewNormalError(NormalSubcategoryDataManagement, 6, http.StatusInternalServerError, "failed to create backup")
	ErrBackupRestoreFailed            = NewNormalError(NormalSubcategoryDataManagement, 7, http.StatusInternalServerError, "failed to restore backup")
	ErrBackupInvalidFileName          = NewNormalError(NormalSubcategoryDataManagement, 8, http.StatusBadRequest, "invalid backup file name")
	ErrBackupDeleteFailed             = NewNormalError(NormalSubcategoryDataManagement, 9, http.StatusInternalServerError, "failed to delete backup")
	ErrBackupDatabaseTypeNotSupported = NewNormalError(NormalSubcategoryDataManagement, 10, http.StatusBadRequest, "backup is only supported for SQLite databases")
)
