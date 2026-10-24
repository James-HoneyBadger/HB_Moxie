package services

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/James-HoneyBadger/ratel-ledger/pkg/core"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/datastore"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/models"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/settings"
)

func TestGetAccountsAndCategoriesTotalInflowAndOutflow_SQLAggregation(t *testing.T) {
	workingDir := t.TempDir()
	databaseDir := filepath.Join(workingDir, "data")

	assert.Nil(t, os.MkdirAll(databaseDir, 0750))

	config := &settings.Config{
		WorkingPath: workingDir,
		DatabaseConfig: &settings.DatabaseConfig{
			DatabaseType:      settings.Sqlite3DbType,
			DatabasePath:      filepath.Join(databaseDir, "ratel-ledger.db"),
			MaxIdleConnection: 1,
			MaxOpenConnection: 1,
		},
	}

	previousConfig := settings.Container.GetCurrentConfig()
	settings.SetCurrentConfig(config)
	assert.Nil(t, datastore.InitializeDataStore(config))

	defer func() {
		_ = datastore.Container.Close()
		settings.SetCurrentConfig(previousConfig)
	}()

	assert.Nil(t, datastore.Container.UserDataStore.SyncStructs(new(models.Transaction)))

	ctx := core.NewNullContext()
	database := datastore.Container.UserDataStore.Choose(0)
	transactionTime := time.Now().Unix() * 1000

	insertTestTransaction(t, database, ctx, 1, models.TRANSACTION_DB_TYPE_INCOME, 1, 1, 0, 100, transactionTime)
	insertTestTransaction(t, database, ctx, 2, models.TRANSACTION_DB_TYPE_INCOME, 1, 1, 0, 50, transactionTime+1)
	insertTestTransaction(t, database, ctx, 3, models.TRANSACTION_DB_TYPE_EXPENSE, 2, 2, 0, 30, transactionTime+2)
	insertTestTransaction(t, database, ctx, 4, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, 0, 1, 2, 70, transactionTime+3)
	insertTestTransaction(t, database, ctx, 5, models.TRANSACTION_DB_TYPE_TRANSFER_IN, 0, 2, 1, 70, transactionTime+4)

	rangeStart := time.Now().Unix() - 3600
	rangeEnd := time.Now().Unix() + 3600

	totalAmounts, err := Transactions.GetAccountsAndCategoriesTotalInflowAndOutflow(ctx, 1, rangeStart, rangeEnd, nil, false, "", time.UTC, false)
	assert.Nil(t, err)
	assert.Len(t, totalAmounts, 4)

	incomeAmounts, expenseAmounts, err := Transactions.GetAccountsTotalIncomeAndExpense(ctx, 1, rangeStart, rangeEnd, nil, nil, time.UTC, false)
	assert.Nil(t, err)
	assert.Equal(t, int64(150), incomeAmounts[1])
	assert.Equal(t, int64(30), expenseAmounts[2])

	amountsByKey := make(map[string]int64)

	for _, totalAmount := range totalAmounts {
		key := transactionStatisticsGroupKey(totalAmount)
		amountsByKey[key] += totalAmount.Amount
	}

	assert.Equal(t, int64(150), amountsByKey["2_1_1_0"])
	assert.Equal(t, int64(30), amountsByKey["3_2_2_0"])
	assert.Equal(t, int64(70), amountsByKey["4_0_1_2"])
	assert.Equal(t, int64(70), amountsByKey["5_0_2_1"])
}

func insertTestTransaction(t *testing.T, database *datastore.Database, ctx core.Context, transactionId int64, transactionType models.TransactionDbType, categoryId int64, accountId int64, relatedAccountId int64, amount int64, transactionTime int64) {
	t.Helper()

	err := database.Exec(ctx, "INSERT INTO `transaction` (transaction_id, uid, deleted, type, category_id, account_id, transaction_time, timezone_utc_offset, amount, related_id, related_account_id, related_account_amount, hide_amount, comment) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		transactionId, 1, false, transactionType, categoryId, accountId, transactionTime, 0, amount, 0, relatedAccountId, 0, false, "")

	assert.Nil(t, err)
}

func transactionStatisticsGroupKey(transaction *models.Transaction) string {
	return fmt.Sprintf("%d_%d_%d_%d", transaction.Type, transaction.CategoryId, transaction.AccountId, transaction.RelatedAccountId)
}
