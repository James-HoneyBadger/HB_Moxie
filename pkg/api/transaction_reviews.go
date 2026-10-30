package api

import (
	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/errs"
	"github.com/James-HoneyBadger/honey-ledger/pkg/log"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/services"
)

// TransactionReviewsApi represents the transaction review api
type TransactionReviewsApi struct {
	reviews *services.TransactionReviewService
}

// Initialize a transaction review api singleton instance
var (
	TransactionReviews = &TransactionReviewsApi{
		reviews: services.TransactionReviews,
	}
)

// TransactionReviewHandler returns duplicates, unusual spending and detected
// subscriptions for the current user.
func (a *TransactionReviewsApi) TransactionReviewHandler(c *core.WebContext) (any, *errs.Error) {
	var reviewReq models.TransactionReviewRequest
	err := c.ShouldBindQuery(&reviewReq)

	if err != nil {
		log.Warnf(c, "[transaction_reviews.TransactionReviewHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	reviewResp, err := a.reviews.Review(c, uid, &reviewReq)

	if err != nil {
		log.Errorf(c, "[transaction_reviews.TransactionReviewHandler] failed to review transactions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return reviewResp, nil
}

// TransactionReviewDismissHandler hides a finding the user does not want to see again.
func (a *TransactionReviewsApi) TransactionReviewDismissHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.TransactionReviewDismissRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Warnf(c, "[transaction_reviews.TransactionReviewDismissHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	if err := a.reviews.Dismiss(c, uid, &request); err != nil {
		log.Errorf(c, "[transaction_reviews.TransactionReviewDismissHandler] failed to dismiss finding for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return true, nil
}

// TransactionReviewRestoreHandler shows a dismissed finding again.
func (a *TransactionReviewsApi) TransactionReviewRestoreHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.TransactionReviewDismissRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Warnf(c, "[transaction_reviews.TransactionReviewRestoreHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	if err := a.reviews.Restore(c, uid, &request); err != nil {
		log.Errorf(c, "[transaction_reviews.TransactionReviewRestoreHandler] failed to restore finding for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return true, nil
}
