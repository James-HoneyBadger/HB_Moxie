package api

import (
	"github.com/James-HoneyBadger/ratel-ledger/pkg/core"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/errs"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/log"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/models"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/services"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/settings"
)

// BackupsApi represents backup api
type BackupsApi struct {
	ApiUsingConfig
	users   *services.UserService
	backups *services.BackupService
}

// Initialize a backup api singleton instance
var (
	Backups = &BackupsApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		users:   services.Users,
		backups: services.Backups,
	}
)

// BackupListHandler returns all backup files
func (a *BackupsApi) BackupListHandler(c *core.WebContext) (any, *errs.Error) {
	backups, err := a.backups.ListBackups(c)

	if err != nil {
		log.Errorf(c, "[backups.BackupListHandler] failed to list backups, because %s", err.Error())
		return nil, errs.ErrOperationFailed
	}

	if backups == nil {
		backups = []*models.BackupListResponse{}
	}

	return backups, nil
}

// BackupCreateHandler creates a new backup
func (a *BackupsApi) BackupCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var backupReq models.BackupRequest
	err := c.ShouldBindJSON(&backupReq)

	if err != nil {
		log.Warnf(c, "[backups.BackupCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Warnf(c, "[backups.BackupCreateHandler] failed to get user for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if !a.users.IsPasswordEqualsUserPassword(backupReq.Password, user) {
		return nil, errs.ErrUserPasswordWrong
	}

	backup, err := a.backups.CreateBackup(c)

	if err != nil {
		if errs.IsCustomError(err) {
			return nil, errs.Or(err, errs.ErrBackupCreationFailed)
		}

		log.Errorf(c, "[backups.BackupCreateHandler] failed to create backup for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrBackupCreationFailed
	}

	log.Infof(c, "[backups.BackupCreateHandler] user \"uid:%d\" created backup \"%s\"", uid, backup.FileName)
	return backup, nil
}

// BackupDownloadHandler downloads a backup file
func (a *BackupsApi) BackupDownloadHandler(c *core.WebContext) ([]byte, string, *errs.Error) {
	var downloadReq models.BackupDownloadRequest
	err := c.ShouldBindQuery(&downloadReq)

	if err != nil {
		log.Warnf(c, "[backups.BackupDownloadHandler] parse request failed, because %s", err.Error())
		return nil, "", errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	content, err := a.backups.GetBackupFileContent(c, downloadReq.FileName)

	if err != nil {
		if errs.IsCustomError(err) {
			return nil, "", errs.Or(err, errs.ErrBackupNotFound)
		}

		log.Errorf(c, "[backups.BackupDownloadHandler] failed to read backup file \"%s\", because %s", downloadReq.FileName, err.Error())
		return nil, "", errs.ErrOperationFailed
	}

	uid := c.GetCurrentUid()
	log.Infof(c, "[backups.BackupDownloadHandler] user \"uid:%d\" downloaded backup \"%s\"", uid, downloadReq.FileName)

	return content, downloadReq.FileName, nil
}

// BackupRestoreHandler restores from a backup
func (a *BackupsApi) BackupRestoreHandler(c *core.WebContext) (any, *errs.Error) {
	var restoreReq models.BackupRestoreRequest
	err := c.ShouldBindJSON(&restoreReq)

	if err != nil {
		log.Warnf(c, "[backups.BackupRestoreHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Warnf(c, "[backups.BackupRestoreHandler] failed to get user for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if !a.users.IsPasswordEqualsUserPassword(restoreReq.Password, user) {
		return nil, errs.ErrUserPasswordWrong
	}

	err = a.backups.RestoreBackup(c, restoreReq.FileName)

	if err != nil {
		if errs.IsCustomError(err) {
			return nil, errs.Or(err, errs.ErrBackupRestoreFailed)
		}

		log.Errorf(c, "[backups.BackupRestoreHandler] failed to restore backup \"%s\" for user \"uid:%d\", because %s", restoreReq.FileName, uid, err.Error())
		return nil, errs.ErrBackupRestoreFailed
	}

	log.Infof(c, "[backups.BackupRestoreHandler] user \"uid:%d\" restored backup \"%s\"", uid, restoreReq.FileName)
	return true, nil
}

// BackupDeleteHandler deletes a backup file
func (a *BackupsApi) BackupDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var deleteReq models.BackupDeleteRequest
	err := c.ShouldBindJSON(&deleteReq)

	if err != nil {
		log.Warnf(c, "[backups.BackupDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	err = a.backups.DeleteBackup(c, deleteReq.FileName)

	if err != nil {
		if errs.IsCustomError(err) {
			return nil, errs.Or(err, errs.ErrBackupNotFound)
		}

		log.Errorf(c, "[backups.BackupDeleteHandler] failed to delete backup \"%s\" for user \"uid:%d\", because %s", deleteReq.FileName, uid, err.Error())
		return nil, errs.ErrBackupDeleteFailed
	}

	log.Infof(c, "[backups.BackupDeleteHandler] user \"uid:%d\" deleted backup \"%s\"", uid, deleteReq.FileName)
	return true, nil
}
