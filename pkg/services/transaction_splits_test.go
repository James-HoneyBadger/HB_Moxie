package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
)

func TestNormalizeTransactionSplits(t *testing.T) {
	t.Run("trims names and keeps amounts", func(t *testing.T) {
		normalized, err := normalizeTransactionSplits([]models.TransactionSplitRequestItem{
			{PersonName: "  Alice  ", Amount: 1200},
			{PersonName: "Bob", Amount: 800},
		})

		assert.Nil(t, err)
		assert.Equal(t, []models.TransactionSplitRequestItem{
			{PersonName: "Alice", Amount: 1200},
			{PersonName: "Bob", Amount: 800},
		}, normalized)
	})

	rejected := map[string][]models.TransactionSplitRequestItem{
		"empty name":       {{PersonName: "   ", Amount: 100}},
		"oversized name":   {{PersonName: string(make([]rune, models.TransactionSplitMaxPersonNameLength+1)), Amount: 100}},
		"zero amount":      {{PersonName: "Alice", Amount: 0}},
		"negative amount":  {{PersonName: "Alice", Amount: -100}},
		"duplicate names":  {{PersonName: "Alice", Amount: 100}, {PersonName: " alice ", Amount: 100}},
		"too many entries": manySplitItems(models.TransactionSplitMaxCount + 1),
	}

	for name, items := range rejected {
		t.Run("rejects "+name, func(t *testing.T) {
			_, err := normalizeTransactionSplits(items)
			assert.NotNil(t, err)
		})
	}
}

func TestSummarizeSharedExpenses(t *testing.T) {
	splits := []*models.TransactionSplit{
		{TransactionId: 1, PersonName: "Alice", Currency: "USD", Amount: 500, Direction: models.SPLIT_DIRECTION_THEY_OWE_ME, CreatedUnixTime: 100},
		{TransactionId: 2, PersonName: "Alice", Currency: "USD", Amount: 250, Direction: models.SPLIT_DIRECTION_THEY_OWE_ME, CreatedUnixTime: 300},
		{TransactionId: 3, PersonName: "Bob", Currency: "USD", Amount: 900, Direction: models.SPLIT_DIRECTION_THEY_OWE_ME, CreatedUnixTime: 200},
		{TransactionId: 4, PersonName: "Alice", Currency: "EUR", Amount: 100, Direction: models.SPLIT_DIRECTION_THEY_OWE_ME, CreatedUnixTime: 400},
		// Deleted transaction: must not count.
		{TransactionId: 5, PersonName: "Carol", Currency: "USD", Amount: 700, Direction: models.SPLIT_DIRECTION_THEY_OWE_ME, CreatedUnixTime: 500},
	}

	summary := summarizeSharedExpenses(splits, nil, map[int64]bool{5: true})

	assert.Len(t, summary, 3)

	// Sorted by currency, then by amount descending.
	assert.Equal(t, "Alice", summary[0].PersonName)
	assert.Equal(t, "EUR", summary[0].Currency)
	assert.Equal(t, int64(100), summary[0].Total)

	assert.Equal(t, "Bob", summary[1].PersonName)
	assert.Equal(t, "USD", summary[1].Currency)
	assert.Equal(t, int64(900), summary[1].Total)

	assert.Equal(t, "Alice", summary[2].PersonName)
	assert.Equal(t, int64(750), summary[2].Total)
	assert.Equal(t, int32(2), summary[2].EntryCount)
	assert.Equal(t, int64(300), summary[2].LastTime)
	assert.Equal(t, int64(750), summary[2].OwedToMe)
	assert.Equal(t, int64(0), summary[2].IOwe)
}

func TestSummarizeSharedExpensesNetsDirectionsAndSettlements(t *testing.T) {
	splits := []*models.TransactionSplit{
		{TransactionId: 1, PersonName: "Priya", Currency: "USD", Amount: 5000, Direction: models.SPLIT_DIRECTION_THEY_OWE_ME, CreatedUnixTime: 100},
		{TransactionId: 2, PersonName: "Priya", Currency: "USD", Amount: 2000, Direction: models.SPLIT_DIRECTION_I_OWE_THEM, CreatedUnixTime: 200},
	}

	settlements := []*models.SharedExpenseSettlement{
		{PersonName: "Priya", Currency: "USD", Amount: 1000, CreatedUnixTime: 300},
	}

	summary := summarizeSharedExpenses(splits, settlements, nil)

	assert.Len(t, summary, 1)
	assert.Equal(t, int64(5000), summary[0].OwedToMe)
	assert.Equal(t, int64(2000), summary[0].IOwe)
	assert.Equal(t, int64(1000), summary[0].Settled)
	assert.Equal(t, int64(2000), summary[0].Total, "owed to me minus I owe minus settlements")
	assert.Equal(t, int64(300), summary[0].LastTime)
}

func TestSummarizeSharedExpensesHidesSettledPeople(t *testing.T) {
	splits := []*models.TransactionSplit{
		{TransactionId: 1, PersonName: "Sam", Currency: "USD", Amount: 3000, Direction: models.SPLIT_DIRECTION_THEY_OWE_ME, CreatedUnixTime: 100},
	}

	settlements := []*models.SharedExpenseSettlement{
		{PersonName: "Sam", Currency: "USD", Amount: 3000},
	}

	assert.Empty(t, summarizeSharedExpenses(splits, settlements, nil), "a fully settled person needs no tracking")
}

func TestSummarizeSharedExpensesKeepsDebtsOwedByUser(t *testing.T) {
	splits := []*models.TransactionSplit{
		{TransactionId: 1, PersonName: "Ines", Currency: "USD", Amount: 4200, Direction: models.SPLIT_DIRECTION_I_OWE_THEM, CreatedUnixTime: 100},
	}

	summary := summarizeSharedExpenses(splits, nil, nil)

	assert.Len(t, summary, 1)
	assert.Equal(t, int64(-4200), summary[0].Total, "negative means the user owes the person")
}

func TestSummarizeSharedExpensesWithoutSplits(t *testing.T) {
	assert.Empty(t, summarizeSharedExpenses(nil, nil, nil))
}

func manySplitItems(count int) []models.TransactionSplitRequestItem {
	items := make([]models.TransactionSplitRequestItem, 0, count)

	for i := 0; i < count; i++ {
		items = append(items, models.TransactionSplitRequestItem{
			PersonName: "person" + string(rune('a'+i)),
			Amount:     100,
		})
	}

	return items
}

func TestBuildSharedTransactionSubCondition(t *testing.T) {
	service := &TransactionService{}

	t.Run("shared keeps split transactions only", func(t *testing.T) {
		condition, params := service.buildSharedTransactionSubCondition(7, models.TRANSACTION_SHARED_FILTER_SHARED)

		assert.Contains(t, condition, "IN (SELECT transaction_id FROM transaction_split")
		assert.NotContains(t, condition, "NOT IN")
		assert.Equal(t, []any{int64(7)}, params)
	})

	t.Run("unshared excludes split transactions", func(t *testing.T) {
		condition, params := service.buildSharedTransactionSubCondition(7, models.TRANSACTION_SHARED_FILTER_UNSHARED)

		assert.Contains(t, condition, "NOT IN (SELECT transaction_id FROM transaction_split")
		assert.Equal(t, []any{int64(7)}, params)
	})

	t.Run("no filter adds no condition", func(t *testing.T) {
		condition, params := service.buildSharedTransactionSubCondition(7, "")

		assert.Empty(t, condition)
		assert.Empty(t, params)
	})
}
