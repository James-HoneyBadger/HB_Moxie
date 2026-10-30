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
		{TransactionId: 1, PersonName: "Alice", Currency: "USD", Amount: 500, CreatedUnixTime: 100},
		{TransactionId: 2, PersonName: "Alice", Currency: "USD", Amount: 250, CreatedUnixTime: 300},
		{TransactionId: 3, PersonName: "Bob", Currency: "USD", Amount: 900, CreatedUnixTime: 200},
		{TransactionId: 4, PersonName: "Alice", Currency: "EUR", Amount: 100, CreatedUnixTime: 400},
		// Deleted transaction: must not count.
		{TransactionId: 5, PersonName: "Carol", Currency: "USD", Amount: 700, CreatedUnixTime: 500},
	}

	summary := summarizeSharedExpenses(splits, map[int64]bool{5: true})

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
}

func TestSummarizeSharedExpensesWithoutSplits(t *testing.T) {
	assert.Empty(t, summarizeSharedExpenses(nil, nil))
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
