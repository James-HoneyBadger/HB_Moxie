package api

import (
	"strings"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/errs"
	"github.com/James-HoneyBadger/honey-ledger/pkg/log"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/services"
	"github.com/James-HoneyBadger/honey-ledger/pkg/utils"
)

// TransactionSplitsApi represents the shared expense api
type TransactionSplitsApi struct {
	splits *services.TransactionSplitService
}

// Initialize a transaction split api singleton instance
var (
	TransactionSplits = &TransactionSplitsApi{
		splits: services.TransactionSplits,
	}
)

// TransactionSplitListHandler returns the splits recorded for one transaction.
func (a *TransactionSplitsApi) TransactionSplitListHandler(c *core.WebContext) (any, *errs.Error) {
	transactionId, err := utils.StringToInt64(c.Query("transactionId"))

	if err != nil || transactionId <= 0 {
		log.Warnf(c, "[transaction_splits.TransactionSplitListHandler] transaction id is invalid")
		return nil, errs.ErrTransactionNotFound
	}

	uid := c.GetCurrentUid()
	splits, err := a.splits.GetTransactionSplits(c, uid, transactionId)

	if err != nil {
		log.Errorf(c, "[transaction_splits.TransactionSplitListHandler] failed to get splits of transaction \"id:%d\" for user \"uid:%d\", because %s", transactionId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return splits, nil
}

// TransactionSplitUpdateHandler replaces the splits of one transaction.
func (a *TransactionSplitsApi) TransactionSplitUpdateHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.TransactionSplitUpdateRequest
	err := c.ShouldBindJSON(&request)

	if err != nil {
		log.Warnf(c, "[transaction_splits.TransactionSplitUpdateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	splits, err := a.splits.SetTransactionSplits(c, uid, &request)

	if err != nil {
		log.Errorf(c, "[transaction_splits.TransactionSplitUpdateHandler] failed to update splits of transaction \"id:%d\" for user \"uid:%d\", because %s", request.TransactionId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return splits, nil
}

// SharedExpenseSummaryHandler returns what each person owes the current user.
func (a *TransactionSplitsApi) SharedExpenseSummaryHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	summary, err := a.splits.GetSharedExpenseSummary(c, uid)

	if err != nil {
		log.Errorf(c, "[transaction_splits.SharedExpenseSummaryHandler] failed to summarize shared expenses for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return summary, nil
}

// TransactionSplitIndexHandler returns the split state of the transactions in a time range.
func (a *TransactionSplitsApi) TransactionSplitIndexHandler(c *core.WebContext) (any, *errs.Error) {
	minTime, _ := utils.StringToInt64(c.Query("min_time"))
	maxTime, _ := utils.StringToInt64(c.Query("max_time"))

	uid := c.GetCurrentUid()
	index, err := a.splits.GetTransactionSplitIndex(c, uid, minTime, maxTime)

	if err != nil {
		log.Errorf(c, "[transaction_splits.TransactionSplitIndexHandler] failed to get split index for user uid:%d, because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return index, nil
}

// SharedExpensePersonHandler returns the individual shared expenses owed by one person.
func (a *TransactionSplitsApi) SharedExpensePersonHandler(c *core.WebContext) (any, *errs.Error) {
	personName := strings.TrimSpace(c.Query("personName"))

	if personName == "" {
		log.Warnf(c, "[transaction_splits.SharedExpensePersonHandler] person name is invalid")
		return nil, errs.ErrTransactionSplitPersonNameInvalid
	}

	uid := c.GetCurrentUid()
	entries, err := a.splits.GetSharedExpensesByPerson(c, uid, personName)

	if err != nil {
		log.Errorf(c, "[transaction_splits.SharedExpensePersonHandler] failed to get shared expenses of person %s for user uid:%d, because %s", personName, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return entries, nil
}
