package api

import (
	"github.com/James-HoneyBadger/ratel-ledger/pkg/core"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/errs"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/log"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/models"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/services"
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
	reviewResp, err := a.reviews.Review(c, uid, reviewReq.Months)

	if err != nil {
		log.Errorf(c, "[transaction_reviews.TransactionReviewHandler] failed to review transactions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return reviewResp, nil
}
