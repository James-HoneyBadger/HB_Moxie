package services

import (
	"time"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/datastore"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/uuid"
)

// TransactionReviewDismissalService stores the review findings a user has
// dismissed, so a known-fine duplicate charge stops reappearing.
type TransactionReviewDismissalService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a transaction review dismissal service singleton instance
var (
	TransactionReviewDismissals = &TransactionReviewDismissalService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetDismissalKeys returns the dismissed findings as "kind\x00key" entries.
func (s *TransactionReviewDismissalService) GetDismissalKeys(c core.Context, uid int64) (map[string]bool, error) {
	session := s.UserDataDB(uid).NewSession(c)
	defer session.Close()

	var dismissals []*models.TransactionReviewDismissal

	if err := session.Where("uid=?", uid).Cols("kind", "key").Find(&dismissals); err != nil {
		return nil, err
	}

	keys := make(map[string]bool, len(dismissals))

	for _, dismissal := range dismissals {
		keys[dismissal.Kind+"\x00"+dismissal.Key] = true
	}

	return keys, nil
}

// AddDismissal records a dismissed finding; dismissing the same finding twice is a
// no-op.
func (s *TransactionReviewDismissalService) AddDismissal(c core.Context, uid int64, kind string, key string) error {
	session := s.UserDataDB(uid).NewSession(c)
	defer session.Close()

	exists, err := session.Where("uid=? AND kind=? AND `key`=?", uid, kind, key).Exist(new(models.TransactionReviewDismissal))

	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	_, err = session.Insert(&models.TransactionReviewDismissal{
		DismissalId:     s.GenerateUuid(uuid.UUID_TYPE_USER),
		Uid:             uid,
		Kind:            kind,
		Key:             key,
		CreatedUnixTime: time.Now().Unix(),
	})

	return err
}

// RemoveDismissal restores a dismissed finding.
func (s *TransactionReviewDismissalService) RemoveDismissal(c core.Context, uid int64, kind string, key string) error {
	_, err := s.UserDataDB(uid).NewSession(c).
		Where("uid=? AND kind=? AND `key`=?", uid, kind, key).
		Delete(new(models.TransactionReviewDismissal))

	return err
}
