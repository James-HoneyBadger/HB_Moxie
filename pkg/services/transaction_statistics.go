package services

import (
	"time"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
)

// TransactionStatisticsRepository is the data-access boundary used by the
// transaction statistics service. Keeping it as an interface allows the
// aggregation logic to be tested independently from xorm.
type TransactionStatisticsRepository interface {
	GetAccountsTotalIncomeAndExpense(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, excludeAccountIds []int64, excludeCategoryIds []int64, clientTimezone *time.Location, useTransactionTimezone bool) (map[int64]int64, map[int64]int64, error)
	GetAccountsAndCategoriesTotalInflowAndOutflow(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, tagFilters []*models.TransactionTagFilter, noTags bool, keyword string, clientTimezone *time.Location, useTransactionTimezone bool) ([]*models.Transaction, error)
	GetAccountsAndCategoriesMonthlyInflowAndOutflow(c core.Context, uid int64, startYear int32, startMonth int32, endYear int32, endMonth int32, tagFilters []*models.TransactionTagFilter, noTags bool, keyword string, clientTimezone *time.Location, useTransactionTimezone bool) (map[int32][]*models.Transaction, error)
	GetAllAccountsDailyOpeningAndClosingBalance(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, clientTimezone *time.Location) (map[int32][]*models.TransactionWithAccountBalance, error)
}

// TransactionStatisticsService provides transaction aggregation for the
// statistics, trends and overview APIs.
type TransactionStatisticsService struct {
	repository TransactionStatisticsRepository
}

// NewTransactionStatisticsService returns a statistics service backed by the given repository.
func NewTransactionStatisticsService(repository TransactionStatisticsRepository) *TransactionStatisticsService {
	return &TransactionStatisticsService{
		repository: repository,
	}
}

// TransactionStatistics is the application-wide transaction statistics service.
var TransactionStatistics = NewTransactionStatisticsService(Transactions)

// GetAccountsTotalIncomeAndExpense returns income and expense totals per account.
func (s *TransactionStatisticsService) GetAccountsTotalIncomeAndExpense(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, excludeAccountIds []int64, excludeCategoryIds []int64, clientTimezone *time.Location, useTransactionTimezone bool) (map[int64]int64, map[int64]int64, error) {
	return s.repository.GetAccountsTotalIncomeAndExpense(c, uid, startUnixTime, endUnixTime, excludeAccountIds, excludeCategoryIds, clientTimezone, useTransactionTimezone)
}

// GetAccountsAndCategoriesTotalInflowAndOutflow returns inflow and outflow totals per account and category.
func (s *TransactionStatisticsService) GetAccountsAndCategoriesTotalInflowAndOutflow(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, tagFilters []*models.TransactionTagFilter, noTags bool, keyword string, clientTimezone *time.Location, useTransactionTimezone bool) ([]*models.Transaction, error) {
	return s.repository.GetAccountsAndCategoriesTotalInflowAndOutflow(c, uid, startUnixTime, endUnixTime, tagFilters, noTags, keyword, clientTimezone, useTransactionTimezone)
}

// GetAccountsAndCategoriesMonthlyInflowAndOutflow returns monthly inflow and outflow totals per account and category.
func (s *TransactionStatisticsService) GetAccountsAndCategoriesMonthlyInflowAndOutflow(c core.Context, uid int64, startYear int32, startMonth int32, endYear int32, endMonth int32, tagFilters []*models.TransactionTagFilter, noTags bool, keyword string, clientTimezone *time.Location, useTransactionTimezone bool) (map[int32][]*models.Transaction, error) {
	return s.repository.GetAccountsAndCategoriesMonthlyInflowAndOutflow(c, uid, startYear, startMonth, endYear, endMonth, tagFilters, noTags, keyword, clientTimezone, useTransactionTimezone)
}

// GetAllAccountsDailyOpeningAndClosingBalance returns daily balances for all accounts.
func (s *TransactionStatisticsService) GetAllAccountsDailyOpeningAndClosingBalance(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, clientTimezone *time.Location) (map[int32][]*models.TransactionWithAccountBalance, error) {
	return s.repository.GetAllAccountsDailyOpeningAndClosingBalance(c, uid, maxTransactionTime, minTransactionTime, clientTimezone)
}
