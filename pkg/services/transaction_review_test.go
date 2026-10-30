package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/utils"
)

const reviewTestDay = int64(24 * 60 * 60)
const reviewTestBaseTime = int64(1700000000)

func reviewTestTransaction(id int64, dayOffset int64, txnType models.TransactionDbType, amount int64, accountId int64, categoryId int64, comment string) *models.Transaction {
	return &models.Transaction{
		TransactionId:   id,
		Type:            txnType,
		Amount:          amount,
		AccountId:       accountId,
		CategoryId:      categoryId,
		Comment:         comment,
		TransactionTime: utils.GetMinTransactionTimeFromUnixTime(reviewTestBaseTime + dayOffset*reviewTestDay),
	}
}

func TestDetectDuplicateTransactions(t *testing.T) {
	testCases := []struct {
		name           string
		transactions   []*models.Transaction
		expectedGroups int
		expectedSizes  []int
	}{
		{
			name: "same amount one day apart is a duplicate",
			transactions: []*models.Transaction{
				reviewTestTransaction(1, 0, models.TRANSACTION_DB_TYPE_EXPENSE, 1250, 10, 20, "Coffee"),
				reviewTestTransaction(2, 1, models.TRANSACTION_DB_TYPE_EXPENSE, 1250, 10, 20, "Coffee"),
			},
			expectedGroups: 1,
			expectedSizes:  []int{2},
		},
		{
			name: "same amount outside the window is not a duplicate",
			transactions: []*models.Transaction{
				reviewTestTransaction(1, 0, models.TRANSACTION_DB_TYPE_EXPENSE, 1250, 10, 20, "Coffee"),
				reviewTestTransaction(2, 10, models.TRANSACTION_DB_TYPE_EXPENSE, 1250, 10, 20, "Coffee"),
			},
			expectedGroups: 0,
		},
		{
			name: "same amount in different accounts is not a duplicate",
			transactions: []*models.Transaction{
				reviewTestTransaction(1, 0, models.TRANSACTION_DB_TYPE_EXPENSE, 1250, 10, 20, "Coffee"),
				reviewTestTransaction(2, 1, models.TRANSACTION_DB_TYPE_EXPENSE, 1250, 11, 20, "Coffee"),
			},
			expectedGroups: 0,
		},
		{
			name: "transfers are ignored",
			transactions: []*models.Transaction{
				reviewTestTransaction(1, 0, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, 5000, 10, 0, ""),
				reviewTestTransaction(2, 1, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, 5000, 10, 0, ""),
			},
			expectedGroups: 0,
		},
		{
			name: "a repeated daily amount is a habit, not a duplicate",
			transactions: []*models.Transaction{
				reviewTestTransaction(1, 0, models.TRANSACTION_DB_TYPE_EXPENSE, 800, 10, 20, "Bus"),
				reviewTestTransaction(2, 1, models.TRANSACTION_DB_TYPE_EXPENSE, 800, 10, 20, "Bus"),
				reviewTestTransaction(3, 2, models.TRANSACTION_DB_TYPE_EXPENSE, 800, 10, 20, "Bus"),
				reviewTestTransaction(4, 3, models.TRANSACTION_DB_TYPE_EXPENSE, 800, 10, 20, "Bus"),
				reviewTestTransaction(5, 4, models.TRANSACTION_DB_TYPE_EXPENSE, 800, 10, 20, "Bus"),
				reviewTestTransaction(6, 5, models.TRANSACTION_DB_TYPE_EXPENSE, 800, 10, 20, "Bus"),
			},
			expectedGroups: 0,
		},
		{
			name: "a chain within the window becomes one group",
			transactions: []*models.Transaction{
				reviewTestTransaction(1, 0, models.TRANSACTION_DB_TYPE_EXPENSE, 800, 10, 20, "Bus"),
				reviewTestTransaction(2, 1, models.TRANSACTION_DB_TYPE_EXPENSE, 800, 10, 20, "Bus"),
				reviewTestTransaction(3, 2, models.TRANSACTION_DB_TYPE_EXPENSE, 800, 10, 20, "Bus"),
			},
			expectedGroups: 1,
			expectedSizes:  []int{3},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			groups := detectDuplicateTransactions(testCase.transactions)

			assert.Len(t, groups, testCase.expectedGroups)

			for i, size := range testCase.expectedSizes {
				assert.Len(t, groups[i].Transactions, size)
			}
		})
	}
}

func TestDetectUnusualTransactions(t *testing.T) {
	transactions := make([]*models.Transaction, 0, 12)
	transactionId := int64(1)

	// Ten ordinary grocery purchases.
	for i := 0; i < 10; i++ {
		transactions = append(transactions, reviewTestTransaction(transactionId, int64(i), models.TRANSACTION_DB_TYPE_EXPENSE, 10000, 1, 5, "Groceries"))
		transactionId++
	}

	// One far larger purchase in the same category.
	transactions = append(transactions, reviewTestTransaction(transactionId, 11, models.TRANSACTION_DB_TYPE_EXPENSE, 100000, 1, 5, "Groceries"))
	transactionId++

	// A category with too few samples must not produce findings.
	transactions = append(transactions, reviewTestTransaction(transactionId, 12, models.TRANSACTION_DB_TYPE_EXPENSE, 900000, 1, 6, "Rare"))
	transactionId++

	// Income is out of scope.
	transactions = append(transactions, reviewTestTransaction(transactionId, 13, models.TRANSACTION_DB_TYPE_INCOME, 500000, 1, 5, "Salary"))

	unusual := detectUnusualTransactions(transactions)

	assert.Len(t, unusual, 1)
	assert.Equal(t, int64(100000), unusual[0].Transaction.Amount)
	assert.Greater(t, unusual[0].ZScore, models.TransactionReviewUnusualZScoreThreshold)
	assert.Equal(t, int32(11), unusual[0].SampleCount)
}

func TestDetectSubscriptions(t *testing.T) {
	testCases := []struct {
		name             string
		transactions     []*models.Transaction
		expectedCount    int
		expectedCadence  string
		expectedChange   int64
		expectedAverage  int64
		expectedOccurred int32
	}{
		{
			name: "monthly subscription with a price rise",
			transactions: []*models.Transaction{
				reviewTestTransaction(1, 0, models.TRANSACTION_DB_TYPE_EXPENSE, 1000, 1, 7, "Netflix"),
				reviewTestTransaction(2, 30, models.TRANSACTION_DB_TYPE_EXPENSE, 1000, 1, 7, "Netflix"),
				reviewTestTransaction(3, 60, models.TRANSACTION_DB_TYPE_EXPENSE, 1200, 1, 7, "Netflix"),
				reviewTestTransaction(4, 90, models.TRANSACTION_DB_TYPE_EXPENSE, 1200, 1, 7, "Netflix"),
			},
			expectedCount:    1,
			expectedCadence:  models.TransactionReviewCadenceMonthly,
			expectedChange:   200,
			expectedAverage:  1100,
			expectedOccurred: 4,
		},
		{
			name: "weekly subscription",
			transactions: []*models.Transaction{
				reviewTestTransaction(1, 0, models.TRANSACTION_DB_TYPE_EXPENSE, 500, 1, 7, "Gym"),
				reviewTestTransaction(2, 7, models.TRANSACTION_DB_TYPE_EXPENSE, 500, 1, 7, "Gym"),
				reviewTestTransaction(3, 14, models.TRANSACTION_DB_TYPE_EXPENSE, 500, 1, 7, "Gym"),
				reviewTestTransaction(4, 21, models.TRANSACTION_DB_TYPE_EXPENSE, 500, 1, 7, "Gym"),
			},
			expectedCount:    1,
			expectedCadence:  models.TransactionReviewCadenceWeekly,
			expectedChange:   0,
			expectedAverage:  500,
			expectedOccurred: 4,
		},
		{
			name: "irregular repeats are not subscriptions",
			transactions: []*models.Transaction{
				reviewTestTransaction(1, 0, models.TRANSACTION_DB_TYPE_EXPENSE, 700, 1, 7, "Bookshop"),
				reviewTestTransaction(2, 17, models.TRANSACTION_DB_TYPE_EXPENSE, 700, 1, 7, "Bookshop"),
				reviewTestTransaction(3, 45, models.TRANSACTION_DB_TYPE_EXPENSE, 700, 1, 7, "Bookshop"),
			},
			expectedCount: 0,
		},
		{
			name: "fewer than three occurrences is not a subscription",
			transactions: []*models.Transaction{
				reviewTestTransaction(1, 0, models.TRANSACTION_DB_TYPE_EXPENSE, 900, 1, 7, "Cloud"),
				reviewTestTransaction(2, 30, models.TRANSACTION_DB_TYPE_EXPENSE, 900, 1, 7, "Cloud"),
			},
			expectedCount: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			subscriptions := detectSubscriptions(testCase.transactions)

			assert.Len(t, subscriptions, testCase.expectedCount)

			if testCase.expectedCount == 0 {
				return
			}

			assert.Equal(t, testCase.expectedCadence, subscriptions[0].Cadence)
			assert.Equal(t, testCase.expectedChange, subscriptions[0].PriceChange)
			assert.Equal(t, testCase.expectedAverage, subscriptions[0].AverageAmount)
			assert.Equal(t, testCase.expectedOccurred, subscriptions[0].Occurrences)
		})
	}
}

func TestNormalizeMerchantName(t *testing.T) {
	testCases := map[string]string{
		"Netflix":                  "netflix",
		"NETFLIX.COM 1234":         "netflix com",
		"  Spotify   Premium  ":    "spotify premium",
		"ACME-Store #99812":        "acme store",
		"1234 5678":                "",
		"Kaufland Berlin":          "kaufland berlin",
		"Über Eats 0451-2233":      "über eats",
		"Monthly plan for two a b": "monthly plan for two",
	}

	for input, expected := range testCases {
		assert.Equal(t, expected, normalizeMerchantName(input), "input: %q", input)
	}
}

type fakeTransactionReviewRepository struct {
	called       bool
	uid          int64
	minUnixTime  int64
	transactions []*models.Transaction
	err          error
}

func (f *fakeTransactionReviewRepository) GetTransactionsForReview(c core.Context, uid int64, minUnixTime int64) ([]*models.Transaction, error) {
	f.called = true
	f.uid = uid
	f.minUnixTime = minUnixTime

	return f.transactions, f.err
}

func TestTransactionReviewServiceReview(t *testing.T) {
	repository := &fakeTransactionReviewRepository{
		transactions: []*models.Transaction{
			reviewTestTransaction(1, 0, models.TRANSACTION_DB_TYPE_EXPENSE, 1500, 1, 2, "Shop"),
			reviewTestTransaction(2, 1, models.TRANSACTION_DB_TYPE_EXPENSE, 1500, 1, 2, "Shop"),
		},
	}

	service := NewTransactionReviewService(repository)
	response, err := service.Review(core.NewNullContext(), 42, 0)

	assert.Nil(t, err)
	assert.True(t, repository.called)
	assert.Equal(t, int64(42), repository.uid)
	assert.Equal(t, int32(2), response.TotalTransactions)
	assert.Len(t, response.Duplicates, 1)
	assert.Equal(t, int64(12)*30*reviewTestDay, response.EndTime-response.StartTime, "defaults to twelve months")
}

func TestTransactionReviewServiceClampsMonths(t *testing.T) {
	repository := &fakeTransactionReviewRepository{}
	service := NewTransactionReviewService(repository)

	response, err := service.Review(core.NewNullContext(), 42, models.TransactionReviewMaxMonths+100)
	assert.Nil(t, err)
	assert.Equal(t, int64(models.TransactionReviewMaxMonths)*30*reviewTestDay, response.EndTime-response.StartTime)
}
