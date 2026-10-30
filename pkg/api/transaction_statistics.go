package api

import (
	"sort"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/errs"
	"github.com/James-HoneyBadger/honey-ledger/pkg/log"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/utils"
)

// TransactionStatisticsHandler returns transaction statistics of current user
func (a *TransactionsApi) TransactionStatisticsHandler(c *core.WebContext) (any, *errs.Error) {
	var statisticReq models.TransactionStatisticRequest
	err := c.ShouldBindQuery(&statisticReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	noTags := statisticReq.TagFilter == models.TransactionNoTagFilterValue
	var tagFilters []*models.TransactionTagFilter

	if !noTags {
		tagFilters, err = models.ParseTransactionTagFilter(statisticReq.TagFilter)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionStatisticsHandler] parse transaction tag filters error, because %s", err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	uid := c.GetCurrentUid()
	totalAmounts, err := a.transactionStatistics.GetAccountsAndCategoriesTotalInflowAndOutflow(c, uid, statisticReq.StartTime, statisticReq.EndTime, tagFilters, noTags, statisticReq.Keyword, clientTimezone, statisticReq.UseTransactionTimezone)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionStatisticsHandler] failed to get accounts and categories total income and expense for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	statisticResp := &models.TransactionStatisticResponse{
		StartTime: statisticReq.StartTime,
		EndTime:   statisticReq.EndTime,
	}

	statisticResp.Items = make([]*models.TransactionStatisticResponseItem, len(totalAmounts))

	for i := 0; i < len(totalAmounts); i++ {
		totalAmountItem := totalAmounts[i]
		statisticResp.Items[i] = &models.TransactionStatisticResponseItem{
			CategoryId:  totalAmountItem.CategoryId,
			AccountId:   totalAmountItem.AccountId,
			TotalAmount: totalAmountItem.Amount,
		}

		if totalAmountItem.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || totalAmountItem.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			statisticResp.Items[i].RelatedAccountId = totalAmountItem.RelatedAccountId
			statisticResp.Items[i].RelatedAccountType, _ = totalAmountItem.Type.ToTransactionRelatedAccountType()
		}
	}

	return statisticResp, nil
}

// TransactionStatisticsTrendsHandler returns transaction statistics trends of current user
func (a *TransactionsApi) TransactionStatisticsTrendsHandler(c *core.WebContext) (any, *errs.Error) {
	var statisticTrendsReq models.TransactionStatisticTrendsRequest
	err := c.ShouldBindQuery(&statisticTrendsReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsTrendsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsTrendsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	startYear, startMonth, endYear, endMonth, err := statisticTrendsReq.GetNumericYearMonthRange()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsTrendsHandler] cannot parse year month, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	noTags := statisticTrendsReq.TagFilter == models.TransactionNoTagFilterValue
	var tagFilters []*models.TransactionTagFilter

	if !noTags {
		tagFilters, err = models.ParseTransactionTagFilter(statisticTrendsReq.TagFilter)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionStatisticsTrendsHandler] parse transaction tag filters error, because %s", err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	uid := c.GetCurrentUid()
	allMonthlyTotalAmounts, err := a.transactionStatistics.GetAccountsAndCategoriesMonthlyInflowAndOutflow(c, uid, startYear, startMonth, endYear, endMonth, tagFilters, noTags, statisticTrendsReq.Keyword, clientTimezone, statisticTrendsReq.UseTransactionTimezone)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionStatisticsTrendsHandler] failed to get accounts and categories total income and expense for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	statisticTrendsResp := make(models.TransactionStatisticTrendsResponseItemSlice, 0, len(allMonthlyTotalAmounts))

	for yearMonth, monthlyTotalAmounts := range allMonthlyTotalAmounts {
		monthlyStatisticResp := &models.TransactionStatisticTrendsResponseItem{
			Year:  yearMonth / 100,
			Month: yearMonth % 100,
			Items: make([]*models.TransactionStatisticResponseItem, len(monthlyTotalAmounts)),
		}

		for i := 0; i < len(monthlyTotalAmounts); i++ {
			totalAmountItem := monthlyTotalAmounts[i]
			monthlyStatisticResp.Items[i] = &models.TransactionStatisticResponseItem{
				CategoryId:  totalAmountItem.CategoryId,
				AccountId:   totalAmountItem.AccountId,
				TotalAmount: totalAmountItem.Amount,
			}

			if totalAmountItem.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || totalAmountItem.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
				monthlyStatisticResp.Items[i].RelatedAccountId = totalAmountItem.RelatedAccountId
				monthlyStatisticResp.Items[i].RelatedAccountType, _ = totalAmountItem.Type.ToTransactionRelatedAccountType()
			}
		}

		statisticTrendsResp = append(statisticTrendsResp, monthlyStatisticResp)
	}

	sort.Sort(statisticTrendsResp)

	return statisticTrendsResp, nil
}

// TransactionStatisticsAssetTrendsHandler returns transaction statistics asset trends of current user
func (a *TransactionsApi) TransactionStatisticsAssetTrendsHandler(c *core.WebContext) (any, *errs.Error) {
	var statisticAssetTrendsReq models.TransactionStatisticAssetTrendsRequest
	err := c.ShouldBindQuery(&statisticAssetTrendsReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsAssetTrendsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionStatisticsAssetTrendsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()

	maxTransactionTime := int64(0)

	if statisticAssetTrendsReq.EndTime > 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(statisticAssetTrendsReq.EndTime)
	}

	minTransactionTime := int64(0)

	if statisticAssetTrendsReq.StartTime > 0 {
		minTransactionTime = utils.GetMinTransactionTimeFromUnixTime(statisticAssetTrendsReq.StartTime)
	}

	accountDailyBalances, err := a.transactionStatistics.GetAllAccountsDailyOpeningAndClosingBalance(c, uid, maxTransactionTime, minTransactionTime, clientTimezone)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionStatisticsAssetTrendsHandler] failed to get transactions from \"%d\" to \"%d\" for user \"uid:%d\", because %s", statisticAssetTrendsReq.StartTime, statisticAssetTrendsReq.EndTime, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	statisticAssetTrendsResp := make(models.TransactionStatisticAssetTrendsResponseItemSlice, 0)

	for yearMonthDay, dailyAccountBalances := range accountDailyBalances {
		dailyStatisticResp := &models.TransactionStatisticAssetTrendsResponseItem{
			Year:  yearMonthDay / 10000,
			Month: (yearMonthDay % 10000) / 100,
			Day:   yearMonthDay % 100,
			Items: make([]*models.TransactionStatisticAssetTrendsResponseDataItem, len(dailyAccountBalances)),
		}

		for i := 0; i < len(dailyAccountBalances); i++ {
			accountBalance := dailyAccountBalances[i]
			dailyStatisticResp.Items[i] = &models.TransactionStatisticAssetTrendsResponseDataItem{
				AccountId:             accountBalance.AccountId,
				AccountOpeningBalance: accountBalance.AccountOpeningBalance,
				AccountClosingBalance: accountBalance.AccountClosingBalance,
			}
		}

		statisticAssetTrendsResp = append(statisticAssetTrendsResp, dailyStatisticResp)
	}

	sort.Sort(statisticAssetTrendsResp)

	return statisticAssetTrendsResp, nil
}

// TransactionAmountsHandler returns transaction amounts of current user
func (a *TransactionsApi) TransactionAmountsHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionAmountsReq models.TransactionAmountsRequest
	err := c.ShouldBindQuery(&transactionAmountsReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionAmountsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	requestItems, err := transactionAmountsReq.GetTransactionAmountsRequestItems()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionAmountsHandler] get request item failed, because %s", err.Error())
		return nil, errs.ErrQueryItemsInvalid
	}

	if len(requestItems) < 1 {
		log.Warnf(c, "[transactions.TransactionAmountsHandler] parse request failed, because there are no valid items")
		return nil, errs.ErrQueryItemsEmpty
	}

	if len(requestItems) > 20 {
		log.Warnf(c, "[transactions.TransactionAmountsHandler] parse request failed, because there are too many items")
		return nil, errs.ErrQueryItemsTooMuch
	}

	excludeAccountIds := make([]int64, 0)
	excludeCategoryIds := make([]int64, 0)

	if transactionAmountsReq.ExcludeAccountIds != "" {
		excludeAccountIds, err = utils.StringArrayToInt64Array(strings.Split(transactionAmountsReq.ExcludeAccountIds, ","))

		if err != nil {
			return nil, errs.ErrAccountIdInvalid
		}
	}

	if transactionAmountsReq.ExcludeCategoryIds != "" {
		excludeCategoryIds, err = utils.StringArrayToInt64Array(strings.Split(transactionAmountsReq.ExcludeCategoryIds, ","))

		if err != nil {
			return nil, errs.ErrTransactionCategoryIdInvalid
		}
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionAmountsHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()

	accounts, err := a.accounts.GetAllAccountsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionAmountsHandler] failed to get all accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	accountMap := a.accounts.GetAccountMapByList(accounts)

	amountsResp := orderedmap.New[string, *models.TransactionAmountsResponseItem]()

	for i := 0; i < len(requestItems); i++ {
		requestItem := requestItems[i]

		incomeAmounts, expenseAmounts, err := a.transactionStatistics.GetAccountsTotalIncomeAndExpense(c, uid, requestItem.StartTime, requestItem.EndTime, excludeAccountIds, excludeCategoryIds, clientTimezone, transactionAmountsReq.UseTransactionTimezone)

		if err != nil {
			log.Errorf(c, "[transactions.TransactionAmountsHandler] failed to get transaction amounts item for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		amountsMap := make(map[string]*models.TransactionAmountsResponseItemAmountInfo)

		for accountId, incomeAmount := range incomeAmounts {
			account, exists := accountMap[accountId]

			if !exists {
				log.Warnf(c, "[transactions.TransactionAmountsHandler] cannot find account for account \"id:%d\" of user \"uid:%d\"", accountId, uid)
				continue
			}

			totalAmounts, exists := amountsMap[account.Currency]

			if !exists {
				totalAmounts = &models.TransactionAmountsResponseItemAmountInfo{
					Currency:      account.Currency,
					IncomeAmount:  0,
					ExpenseAmount: 0,
				}
			}

			totalAmounts.IncomeAmount += incomeAmount
			amountsMap[account.Currency] = totalAmounts
		}

		for accountId, expenseAmount := range expenseAmounts {
			account, exists := accountMap[accountId]

			if !exists {
				log.Warnf(c, "[transactions.TransactionAmountsHandler] cannot find account for account \"id:%d\" of user \"uid:%d\"", accountId, uid)
				continue
			}

			totalAmounts, exists := amountsMap[account.Currency]

			if !exists {
				totalAmounts = &models.TransactionAmountsResponseItemAmountInfo{
					Currency:      account.Currency,
					IncomeAmount:  0,
					ExpenseAmount: 0,
				}
			}

			totalAmounts.ExpenseAmount += expenseAmount
			amountsMap[account.Currency] = totalAmounts
		}

		allTotalAmounts := make(models.TransactionAmountsResponseItemAmountInfoSlice, 0)

		for _, totalAmounts := range amountsMap {
			allTotalAmounts = append(allTotalAmounts, totalAmounts)
		}

		sort.Sort(allTotalAmounts)

		amountsResp.Set(requestItem.Name, &models.TransactionAmountsResponseItem{
			StartTime: requestItem.StartTime,
			EndTime:   requestItem.EndTime,
			Amounts:   allTotalAmounts,
		})
	}

	return amountsResp, nil
}
