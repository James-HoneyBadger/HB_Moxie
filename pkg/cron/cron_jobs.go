package cron

import (
	"time"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/log"
	"github.com/James-HoneyBadger/honey-ledger/pkg/services"
	"github.com/James-HoneyBadger/honey-ledger/pkg/settings"
)

// RemoveExpiredTokensJob represents the cron job which periodically remove expired user tokens from the database
var RemoveExpiredTokensJob = &CronJob{
	Name:        "RemoveExpiredTokens",
	Description: "Periodically remove expired user tokens from the database.",
	Period: CronJobFixedHourPeriod{
		Hour: 0,
	},
	Run: func(c *core.CronContext) error {
		return services.Tokens.DeleteAllExpiredTokens(c)
	},
}

// CreateScheduledTransactionJob represents the cron job which periodically create transaction by scheduled transaction template
var CreateScheduledTransactionJob = &CronJob{
	Name:        "CreateScheduledTransaction",
	Description: "Periodically create transaction by scheduled transaction template.",
	Period: CronJobEvery15MinutesPeriod{
		Second: 0,
	},
	Run: func(c *core.CronContext) error {
		return services.Transactions.CreateScheduledTransactions(c, time.Now().Unix(), c.GetInterval())
	},
}

// NewAutoBackupJob creates a new auto-backup cron job with the given config
func NewAutoBackupJob(config *settings.Config) *CronJob {
	return &CronJob{
		Name:        "AutoBackup",
		Description: "Periodically create backups and prune old ones.",
		Period: CronJobIntervalPeriod{
			Interval: time.Duration(config.AutoBackupIntervalHours) * time.Hour,
		},
		Run: func(c *core.CronContext) error {
			// Create a new backup
			backup, err := services.Backups.CreateBackup(c)

			if err != nil {
				return err
			}

			log.Infof(c, "[cron_jobs.AutoBackup] created backup \"%s\" (%d bytes)", backup.FileName, backup.FileSize)

			// Prune old backups beyond the configured maximum
			backups, err := services.Backups.ListBackups(c)

			if err != nil {
				log.Warnf(c, "[cron_jobs.AutoBackup] failed to list backups for pruning: %s", err.Error())
				return nil // backup was created successfully, pruning failure is non-fatal
			}

			maxCount := int(config.AutoBackupMaxCount)

			if len(backups) > maxCount {
				// backups are sorted newest-first by ListBackups
				toDelete := backups[maxCount:]

				for _, old := range toDelete {
					if delErr := services.Backups.DeleteBackup(c, old.FileName); delErr != nil {
						log.Warnf(c, "[cron_jobs.AutoBackup] failed to delete old backup \"%s\": %s", old.FileName, delErr.Error())
					} else {
						log.Infof(c, "[cron_jobs.AutoBackup] pruned old backup \"%s\"", old.FileName)
					}
				}
			}

			return nil
		},
	}
}
