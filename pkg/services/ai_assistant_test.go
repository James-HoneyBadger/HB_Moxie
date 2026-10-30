package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/settings"
)

func aiAssistantTestSource() aiAssistantDataSource {
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	return aiAssistantDataSource{
		startTime: startTime,
		endTime:   startTime.AddDate(1, 0, 0),
		accounts: []*models.Account{
			{AccountId: 1, Category: models.ACCOUNT_CATEGORY_CHECKING_ACCOUNT, Currency: "USD", Balance: 250000},
			{AccountId: 2, Category: models.ACCOUNT_CATEGORY_CREDIT_CARD, Currency: "USD", Balance: -40000},
			{AccountId: 3, Category: models.ACCOUNT_CATEGORY_CASH, Currency: "EUR", Balance: 10000},
		},
		categories: []*models.TransactionCategory{
			{CategoryId: 10, Name: "Groceries"},
			{CategoryId: 11, Name: "Dining"},
			{CategoryId: 12, Name: "Salary"},
		},
		transactions: []*models.Transaction{
			{Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: 5000, CategoryId: 10, AccountId: 1},
			{Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: 2500, CategoryId: 11, AccountId: 2},
			{Type: models.TRANSACTION_DB_TYPE_INCOME, Amount: 300000, CategoryId: 12, AccountId: 1},
			{Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: 1000, CategoryId: 10, AccountId: 3},
			// Balance modifications are not income or expense and must be ignored.
			{Type: models.TRANSACTION_DB_TYPE_MODIFY_BALANCE, Amount: 999999, AccountId: 1},
			// A transaction on a deleted or unknown account has no currency.
			{Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: 700, CategoryId: 10, AccountId: 99},
		},
		sharedItems: []*models.SharedExpenseSummaryItem{
			{PersonName: "Priya", Currency: "USD", Total: 2500},
			{PersonName: "Zoe", Currency: "USD", Total: -500},
		},
	}
}

func TestAssembleAIAssistantContextTotalsPerCurrency(t *testing.T) {
	context := assembleAIAssistantContext(aiAssistantTestSource())

	assert.Equal(t, "2026-01-01", context.StartTime)
	assert.Equal(t, "2027-01-01", context.EndTime)
	assert.Equal(t, int64(4), context.TransactionCount, "balance modifications and unknown accounts are excluded")

	assert.Len(t, context.Totals, 2)

	eur := context.Totals[0]
	assert.Equal(t, "EUR", eur.Currency)
	assert.Equal(t, int64(0), eur.Income)
	assert.Equal(t, int64(1000), eur.Expense)
	assert.Equal(t, int64(1), eur.Count)

	usd := context.Totals[1]
	assert.Equal(t, "USD", usd.Currency)
	assert.Equal(t, int64(300000), usd.Income)
	assert.Equal(t, int64(7500), usd.Expense)
	assert.Equal(t, int64(3), usd.Count)
}

func TestAssembleAIAssistantContextRanksCategories(t *testing.T) {
	context := assembleAIAssistantContext(aiAssistantTestSource())

	assert.Len(t, context.ExpenseCategories, 3)
	assert.Equal(t, "Groceries", context.ExpenseCategories[0].Category)
	assert.Equal(t, int64(5000), context.ExpenseCategories[0].Amount, "same category across currencies is reported per currency")
	assert.Equal(t, "USD", context.ExpenseCategories[0].Currency)
	assert.Equal(t, int64(2500), context.ExpenseCategories[1].Amount)
	assert.Equal(t, "Groceries", context.ExpenseCategories[2].Category)
	assert.Equal(t, "EUR", context.ExpenseCategories[2].Currency)
}

func TestAssembleAIAssistantContextCapsCategoryRows(t *testing.T) {
	source := aiAssistantTestSource()
	source.categories = nil
	source.transactions = nil

	for index := 0; index < aiAssistantMaxCategoryRows+4; index++ {
		categoryId := int64(100 + index)
		source.categories = append(source.categories, &models.TransactionCategory{CategoryId: categoryId, Name: fmt.Sprintf("Category %d", index)})
		source.transactions = append(source.transactions, &models.Transaction{
			Type: models.TRANSACTION_DB_TYPE_EXPENSE, Amount: int64(100 + index), CategoryId: categoryId, AccountId: 1,
		})
	}

	context := assembleAIAssistantContext(source)

	assert.Len(t, context.ExpenseCategories, aiAssistantMaxCategoryRows)
	assert.Equal(t, int64(111), context.ExpenseCategories[0].Amount, "largest category first")
}

func TestAssembleAIAssistantContextHidesNames(t *testing.T) {
	context := assembleAIAssistantContext(aiAssistantTestSource())

	assert.Len(t, context.Accounts, 3)
	assert.Equal(t, "Account 1", context.Accounts[0].Label)
	assert.Equal(t, "checking account", context.Accounts[0].Category)
	assert.Equal(t, "Account 2", context.Accounts[1].Label)
	assert.Equal(t, "credit card", context.Accounts[1].Category)
	assert.Equal(t, int64(-40000), context.Accounts[1].Balance)

	assert.Len(t, context.SharedExpenses, 2)
	assert.Equal(t, "Person 1", context.SharedExpenses[0].Label)
	assert.Equal(t, int64(2500), context.SharedExpenses[0].Balance)
	assert.Equal(t, int64(-500), context.SharedExpenses[1].Balance, "negative means the user owes the person")

	encoded, err := json.Marshal(context)
	assert.Nil(t, err)
	assert.NotContains(t, string(encoded), "Priya")
	assert.NotContains(t, string(encoded), "Zoe")
	assert.NotContains(t, string(encoded), "Cash Account")
}

func TestAssembleAIAssistantContextWithoutData(t *testing.T) {
	context := assembleAIAssistantContext(aiAssistantDataSource{
		startTime: time.Now(),
		endTime:   time.Now(),
	})

	assert.Equal(t, int64(0), context.TransactionCount)
	assert.Empty(t, context.Totals)
	assert.Empty(t, context.ExpenseCategories)
	assert.Empty(t, context.Accounts)
	assert.Empty(t, context.SharedExpenses)
}

func TestAIAssistantPromptsCarryTheRules(t *testing.T) {
	systemPrompt := buildAIAssistantSystemPrompt()

	assert.Contains(t, systemPrompt, "smallest unit")
	assert.Contains(t, systemPrompt, "Never invent")
	assert.Contains(t, systemPrompt, "Account 1")

	userPrompt := buildAIAssistantUserPrompt("How much did I spend on groceries?", `{"totals":[]}`)
	assert.Contains(t, userPrompt, `{"totals":[]}`)
	assert.Contains(t, userPrompt, "How much did I spend on groceries?")
	assert.True(t, strings.HasPrefix(userPrompt, "Financial summary:"))
}

func TestAIAssistantModelID(t *testing.T) {
	assert.Equal(t, "gpt-test", aiAssistantModelID(&settings.LLMConfig{LLMProvider: settings.OpenAILLMProvider, OpenAIModelID: "gpt-test"}))
	assert.Equal(t, "claude-test", aiAssistantModelID(&settings.LLMConfig{LLMProvider: settings.AnthropicLLMProvider, AnthropicModelID: "claude-test"}))
	assert.Equal(t, "llama", aiAssistantModelID(&settings.LLMConfig{LLMProvider: settings.OllamaLLMProvider, OllamaModelID: "llama"}))
	assert.Equal(t, "", aiAssistantModelID(&settings.LLMConfig{}))
	assert.Equal(t, "", aiAssistantModelID(nil))
}

func TestAIAssistantAccountCategoryName(t *testing.T) {
	assert.Equal(t, "savings account", aiAssistantAccountCategoryName(models.ACCOUNT_CATEGORY_SAVINGS_ACCOUNT))
	assert.Equal(t, "investment", aiAssistantAccountCategoryName(models.ACCOUNT_CATEGORY_INVESTMENT))
	assert.Equal(t, "account", aiAssistantAccountCategoryName(models.ACCOUNT_CATEGORY_DEBT+100))
}

func TestTruncateRunes(t *testing.T) {
	assert.Equal(t, "abc", truncateRunes("abc", 5))
	assert.Equal(t, "abc", truncateRunes("abcdef", 3))
	assert.Equal(t, "日本語", truncateRunes("日本語です", 3), "multi-byte characters are never split")
}

func TestExtractAIAssistantAnswer(t *testing.T) {
	assert.Equal(t, "You spent 75.00 USD.", extractAIAssistantAnswer(`{"answer": "You spent 75.00 USD."}`))
	assert.Equal(t, "plain text", extractAIAssistantAnswer("plain text"), "a non-JSON reply is kept as-is")
	assert.Equal(t, `{"other": 1}`, extractAIAssistantAnswer(`{"other": 1}`), "unrelated JSON is not silently dropped")
	assert.Equal(t, `{"answer": "  "}`, extractAIAssistantAnswer(`{"answer": "  "}`), "an empty answer is not treated as an answer")
}
