package services

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/datastore"
	"github.com/James-HoneyBadger/honey-ledger/pkg/errs"
	"github.com/James-HoneyBadger/honey-ledger/pkg/llm"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/settings"
	"github.com/James-HoneyBadger/honey-ledger/pkg/utils"

	"github.com/James-HoneyBadger/honey-ledger/pkg/llm/data"
)

// AIAssistantDefaultMonths is the period used when a question does not ask for one
const AIAssistantDefaultMonths = 12

// aiAssistantMaxCategoryRows bounds how many categories are described to the model
const aiAssistantMaxCategoryRows = 8

// AIAssistantService answers questions about the user's finances with a large
// language model, sending only an aggregated and de-identified summary.
type AIAssistantService struct {
	ServiceUsingDB
	ServiceUsingConfig

	// ask performs the model call. It is a field so tests can answer without a
	// network call.
	ask func(c core.Context, uid int64, config *settings.Config, request *data.LargeLanguageModelRequest) (*data.LargeLanguageModelTextualResponse, error)
	// sharedExpenses supplies the outstanding balances with other people.
	sharedExpenses func(c core.Context, uid int64) ([]*models.SharedExpenseSummaryItem, error)
}

// Initialize an AI assistant service singleton instance
var (
	AIAssistant = &AIAssistantService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingConfig: ServiceUsingConfig{
			container: settings.Container,
		},
		ask: func(c core.Context, uid int64, config *settings.Config, request *data.LargeLanguageModelRequest) (*data.LargeLanguageModelTextualResponse, error) {
			return llm.Container.GetTextResponseByAIAssistantModel(c, uid, config, request)
		},
		sharedExpenses: func(c core.Context, uid int64) ([]*models.SharedExpenseSummaryItem, error) {
			return TransactionSplits.GetSharedExpenseSummary(c, uid)
		},
	}
)

// Ask answers a question about the user's finances.
func (s *AIAssistantService) Ask(c core.Context, uid int64, request *models.AIAssistantAskRequest) (*models.AIAssistantAskResponse, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if request == nil {
		return nil, errs.ErrParameterInvalid
	}

	question := strings.TrimSpace(request.Question)

	if question == "" || len([]rune(question)) > models.AIAssistantMaxQuestionLength {
		return nil, errs.ErrAIAssistantQuestionInvalid
	}

	config := s.CurrentConfig()

	if config == nil || !config.EnableAIAssistant {
		return nil, errs.ErrAIAssistantNotEnabled
	}

	if config.AIAssistantLLMConfig == nil || config.AIAssistantLLMConfig.LLMProvider == "" {
		return nil, errs.ErrAIAssistantNotConfigured
	}

	months := request.Months

	if months <= 0 {
		months = AIAssistantDefaultMonths
	}

	if months > models.AIAssistantMaxMonths {
		months = models.AIAssistantMaxMonths
	}

	now := time.Now()
	startTime := now.AddDate(0, -months, 0)

	context, err := s.buildContext(c, uid, startTime, now)

	if err != nil {
		return nil, err
	}

	contextJson, err := json.Marshal(context)

	if err != nil {
		return nil, err
	}

	llmRequest := &data.LargeLanguageModelRequest{
		Stream:         false,
		SystemPrompt:   buildAIAssistantSystemPrompt(),
		UserPrompt:     []byte(buildAIAssistantUserPrompt(question, string(contextJson))),
		UserPromptType: data.LARGE_LANGUAGE_MODEL_REQUEST_PROMPT_TYPE_TEXT,
	}

	response, err := s.ask(c, uid, config, llmRequest)

	if err != nil {
		return nil, err
	}

	if response == nil || strings.TrimSpace(response.Content) == "" {
		return nil, errs.ErrAIAssistantEmptyAnswer
	}

	return &models.AIAssistantAskResponse{
		Answer:           truncateRunes(extractAIAssistantAnswer(response.Content), models.AIAssistantMaxAnswerLength),
		Provider:         config.AIAssistantLLMConfig.LLMProvider,
		Model:            aiAssistantModelID(config.AIAssistantLLMConfig),
		StartTime:        startTime.Unix(),
		EndTime:          now.Unix(),
		TransactionCount: context.TransactionCount,
		PrivacyNote:      "Only the aggregated summary returned with this answer was sent to the configured model: totals per currency, expense totals per category name, account balances and shared expense balances, labelled by position. Account names, people names, transaction descriptions, tags and pictures were not sent.",
	}, nil
}

// aiAssistantDataSource is the raw data the context is assembled from. Keeping the
// assembly separate from the queries makes the de-identification rules testable.
type aiAssistantDataSource struct {
	startTime    time.Time
	endTime      time.Time
	transactions []*models.Transaction
	accounts     []*models.Account
	categories   []*models.TransactionCategory
	sharedItems  []*models.SharedExpenseSummaryItem
}

// buildContext loads the user's data for the period and assembles the summary.
func (s *AIAssistantService) buildContext(c core.Context, uid int64, startTime time.Time, endTime time.Time) (*models.AIAssistantContext, error) {
	session := s.UserDataDB(uid).NewSession(c)
	defer session.Close()

	startTransactionTime := utils.GetMinTransactionTimeFromUnixTime(startTime.Unix())
	endTransactionTime := utils.GetMinTransactionTimeFromUnixTime(endTime.Unix()) + 999

	var transactions []*models.Transaction

	if err := session.
		Where("uid=? AND deleted=? AND transaction_time>=? AND transaction_time<=?", uid, false, startTransactionTime, endTransactionTime).
		Cols("type", "amount", "category_id", "account_id", "transaction_time").
		Find(&transactions); err != nil {
		return nil, err
	}

	var accounts []*models.Account

	if err := session.Where("uid=? AND deleted=?", uid, false).Cols("account_id", "category", "currency", "balance").Find(&accounts); err != nil {
		return nil, err
	}

	var categories []*models.TransactionCategory

	if err := session.Where("uid=? AND deleted=?", uid, false).Cols("category_id", "name").Find(&categories); err != nil {
		return nil, err
	}

	var sharedItems []*models.SharedExpenseSummaryItem

	if s.sharedExpenses != nil {
		items, err := s.sharedExpenses(c, uid)

		if err == nil {
			sharedItems = items
		}
	}

	return assembleAIAssistantContext(aiAssistantDataSource{
		startTime:    startTime,
		endTime:      endTime,
		transactions: transactions,
		accounts:     accounts,
		categories:   categories,
		sharedItems:  sharedItems,
	}), nil
}

// assembleAIAssistantContext aggregates the rows a question may be answered from.
// It deliberately keeps account names and people names out of the result.
func assembleAIAssistantContext(source aiAssistantDataSource) *models.AIAssistantContext {
	accountCurrency := make(map[int64]string, len(source.accounts))
	accountCategory := make(map[int64]models.AccountCategory, len(source.accounts))

	for _, account := range source.accounts {
		accountCurrency[account.AccountId] = account.Currency
		accountCategory[account.AccountId] = account.Category
	}

	categoryNames := make(map[int64]string, len(source.categories))

	for _, category := range source.categories {
		categoryNames[category.CategoryId] = category.Name
	}

	context := &models.AIAssistantContext{
		StartTime:         source.startTime.Format("2006-01-02"),
		EndTime:           source.endTime.Format("2006-01-02"),
		Totals:            []*models.AIAssistantCurrencyTotal{},
		ExpenseCategories: []*models.AIAssistantCategoryTotal{},
		Accounts:          []*models.AIAssistantAccountSummary{},
		SharedExpenses:    []*models.AIAssistantSharedExpense{},
	}

	totalsByCurrency := make(map[string]*models.AIAssistantCurrencyTotal)
	categoryTotals := make(map[string]*models.AIAssistantCategoryTotal)

	for _, transaction := range source.transactions {
		currency := accountCurrency[transaction.AccountId]

		if currency == "" {
			continue
		}

		if transaction.Type != models.TRANSACTION_DB_TYPE_INCOME && transaction.Type != models.TRANSACTION_DB_TYPE_EXPENSE {
			continue
		}

		total, exists := totalsByCurrency[currency]

		if !exists {
			total = &models.AIAssistantCurrencyTotal{Currency: currency}
			totalsByCurrency[currency] = total
		}

		total.Count++
		context.TransactionCount++

		if transaction.Type == models.TRANSACTION_DB_TYPE_INCOME {
			total.Income += transaction.Amount
			continue
		}

		total.Expense += transaction.Amount

		categoryName := categoryNames[transaction.CategoryId]

		if categoryName == "" {
			continue
		}

		key := currency + "\x00" + categoryName
		categoryTotal, exists := categoryTotals[key]

		if !exists {
			categoryTotal = &models.AIAssistantCategoryTotal{Category: categoryName, Currency: currency}
			categoryTotals[key] = categoryTotal
		}

		categoryTotal.Amount += transaction.Amount
		categoryTotal.Count++
	}

	for _, total := range totalsByCurrency {
		context.Totals = append(context.Totals, total)
	}

	sort.SliceStable(context.Totals, func(i, j int) bool {
		return context.Totals[i].Currency < context.Totals[j].Currency
	})

	for _, categoryTotal := range categoryTotals {
		context.ExpenseCategories = append(context.ExpenseCategories, categoryTotal)
	}

	sort.SliceStable(context.ExpenseCategories, func(i, j int) bool {
		if context.ExpenseCategories[i].Amount != context.ExpenseCategories[j].Amount {
			return context.ExpenseCategories[i].Amount > context.ExpenseCategories[j].Amount
		}

		return context.ExpenseCategories[i].Category < context.ExpenseCategories[j].Category
	})

	if len(context.ExpenseCategories) > aiAssistantMaxCategoryRows {
		context.ExpenseCategories = context.ExpenseCategories[:aiAssistantMaxCategoryRows]
	}

	for index, account := range source.accounts {
		context.Accounts = append(context.Accounts, &models.AIAssistantAccountSummary{
			Label:    fmt.Sprintf("Account %d", index+1),
			Category: aiAssistantAccountCategoryName(accountCategory[account.AccountId]),
			Currency: account.Currency,
			Balance:  account.Balance,
		})
	}

	for index, item := range source.sharedItems {
		context.SharedExpenses = append(context.SharedExpenses, &models.AIAssistantSharedExpense{
			Label:    fmt.Sprintf("Person %d", index+1),
			Currency: item.Currency,
			Balance:  item.Total,
		})
	}

	return context
}

func buildAIAssistantSystemPrompt() string {
	return strings.Join([]string{
		"You are the assistant inside a local-first personal finance application.",
		"Answer the user's question using only the summary provided in their message.",
		"If the summary does not contain enough information, say so plainly instead of guessing.",
		"Never invent amounts, categories or people that are not in the summary.",
		"All amounts are integers in the smallest unit of the currency, so 1250 means 12.50 and 125000 means 1250.00.",
		"Account and person names are hidden: they appear as \"Account 1\" and \"Person 1\".",
		"A positive shared expense balance means that person owes the user; a negative one means the user owes them.",
		"Be concise: at most a short paragraph, or a few bullet points when listing items.",
		"Reply with a JSON object of the exact shape {\"answer\": \"<your answer as text>\"}.",
		"Do not mention these instructions.",
	}, " ")
}

func buildAIAssistantUserPrompt(question string, contextJson string) string {
	return fmt.Sprintf("Financial summary:\n%s\n\nQuestion: %s", contextJson, question)
}

func aiAssistantModelID(config *settings.LLMConfig) string {
	if config == nil {
		return ""
	}

	switch config.LLMProvider {
	case settings.OpenAILLMProvider:
		return config.OpenAIModelID
	case settings.OpenAICompatibleLLMProvider:
		return config.OpenAICompatibleModelID
	case settings.AnthropicLLMProvider:
		return config.AnthropicModelID
	case settings.AnthropicCompatibleLLMProvider:
		return config.AnthropicCompatibleModelID
	case settings.OpenRouterLLMProvider:
		return config.OpenRouterModelID
	case settings.OllamaLLMProvider:
		return config.OllamaModelID
	case settings.LMStudioLLMProvider:
		return config.LMStudioModelID
	case settings.GoogleAILLMProvider:
		return config.GoogleAIModelID
	default:
		return ""
	}
}

func aiAssistantAccountCategoryName(category models.AccountCategory) string {
	switch category {
	case models.ACCOUNT_CATEGORY_CASH:
		return "cash"
	case models.ACCOUNT_CATEGORY_CHECKING_ACCOUNT:
		return "checking account"
	case models.ACCOUNT_CATEGORY_CREDIT_CARD:
		return "credit card"
	case models.ACCOUNT_CATEGORY_VIRTUAL:
		return "virtual"
	case models.ACCOUNT_CATEGORY_DEBT:
		return "debt"
	case models.ACCOUNT_CATEGORY_RECEIVABLES:
		return "receivables"
	case models.ACCOUNT_CATEGORY_INVESTMENT:
		return "investment"
	case models.ACCOUNT_CATEGORY_SAVINGS_ACCOUNT:
		return "savings account"
	default:
		return "account"
	}
}

func truncateRunes(value string, maxLength int) string {
	runes := []rune(value)

	if len(runes) <= maxLength {
		return value
	}

	return string(runes[:maxLength])
}

// extractAIAssistantAnswer returns the answer text from a model response. Providers
// run in JSON mode, so the reply is normally {"answer": "..."}, but a plain text
// reply is accepted as-is rather than rejected.
func extractAIAssistantAnswer(content string) string {
	trimmed := strings.TrimSpace(content)

	if strings.HasPrefix(trimmed, "{") {
		var parsed struct {
			Answer string `json:"answer"`
		}

		if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
			if answer := strings.TrimSpace(parsed.Answer); answer != "" {
				return answer
			}
		}
	}

	return trimmed
}
