package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
)

type fakeTransactionStatisticsRepository struct {
	totalInflowOutflowCalled bool
}

func (f *fakeTransactionStatisticsRepository) GetAccountsTotalIncomeAndExpense(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, excludeAccountIds []int64, excludeCategoryIds []int64, clientTimezone *time.Location, useTransactionTimezone bool) (map[int64]int64, map[int64]int64, error) {
	return nil, nil, nil
}

func (f *fakeTransactionStatisticsRepository) GetAccountsAndCategoriesTotalInflowAndOutflow(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, tagFilters []*models.TransactionTagFilter, noTags bool, keyword string, clientTimezone *time.Location, useTransactionTimezone bool) ([]*models.Transaction, error) {
	f.totalInflowOutflowCalled = true
	return []*models.Transaction{{Amount: 42}}, nil
}

func (f *fakeTransactionStatisticsRepository) GetAccountsAndCategoriesMonthlyInflowAndOutflow(c core.Context, uid int64, startYear int32, startMonth int32, endYear int32, endMonth int32, tagFilters []*models.TransactionTagFilter, noTags bool, keyword string, clientTimezone *time.Location, useTransactionTimezone bool) (map[int32][]*models.Transaction, error) {
	return nil, nil
}

func (f *fakeTransactionStatisticsRepository) GetAllAccountsDailyOpeningAndClosingBalance(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, clientTimezone *time.Location) (map[int32][]*models.TransactionWithAccountBalance, error) {
	return nil, nil
}

func TestTransactionStatisticsService_DelegatesToRepository(t *testing.T) {
	repository := &fakeTransactionStatisticsRepository{}
	service := NewTransactionStatisticsService(repository)

	result, err := service.GetAccountsAndCategoriesTotalInflowAndOutflow(core.NewNullContext(), 1, 0, 0, nil, false, "", time.UTC, false)

	assert.Nil(t, err)
	assert.True(t, repository.totalInflowOutflowCalled)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(42), result[0].Amount)
}
