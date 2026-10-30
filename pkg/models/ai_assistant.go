package models

// AIAssistantMaxQuestionLength is the maximum length of an AI assistant question
const AIAssistantMaxQuestionLength = 500

// AIAssistantMaxMonths is the longest period a question may cover
const AIAssistantMaxMonths = 60

// AIAssistantMaxAnswerLength is the longest answer kept from a model response
const AIAssistantMaxAnswerLength = 8000

// AIAssistantAskRequest represents all parameters of an AI assistant question
type AIAssistantAskRequest struct {
	Question string `json:"question" binding:"required,max=500"`
	// Months bounds the period the answer may draw on; 0 means the default.
	Months int `json:"months" binding:"min=0,max=60"`
}

// AIAssistantAskResponse represents the answer to a question about the user's data
type AIAssistantAskResponse struct {
	Answer           string `json:"answer"`
	Provider         string `json:"provider"`
	Model            string `json:"model"`
	StartTime        int64  `json:"startTime"`
	EndTime          int64  `json:"endTime"`
	TransactionCount int64  `json:"transactionCount"`
	// PrivacyNote states what was sent to the model, so an answer is never shown
	// without its provenance.
	PrivacyNote string `json:"privacyNote"`
}

// AIAssistantContext is the aggregated, de-identified summary sent to the model.
// Names of accounts and people are never included.
type AIAssistantContext struct {
	StartTime         string                       `json:"startTime"`
	EndTime           string                       `json:"endTime"`
	TransactionCount  int64                        `json:"transactionCount"`
	Totals            []*AIAssistantCurrencyTotal  `json:"totals"`
	ExpenseCategories []*AIAssistantCategoryTotal  `json:"expenseCategories"`
	Accounts          []*AIAssistantAccountSummary `json:"accounts"`
	SharedExpenses    []*AIAssistantSharedExpense  `json:"sharedExpenses"`
}

// AIAssistantCurrencyTotal represents income and expense totals for one currency,
// in the smallest currency unit
type AIAssistantCurrencyTotal struct {
	Currency string `json:"currency"`
	Income   int64  `json:"income"`
	Expense  int64  `json:"expense"`
	Count    int64  `json:"count"`
}

// AIAssistantCategoryTotal represents the expense total of one category
type AIAssistantCategoryTotal struct {
	Category string `json:"category"`
	Currency string `json:"currency"`
	Amount   int64  `json:"amount"`
	Count    int64  `json:"count"`
}

// AIAssistantAccountSummary represents one account balance, labelled by position
// instead of by name
type AIAssistantAccountSummary struct {
	Label    string `json:"label"`
	Category string `json:"category"`
	Currency string `json:"currency"`
	Balance  int64  `json:"balance"`
}

// AIAssistantSharedExpense represents the outstanding balance with one person,
// labelled by position instead of by name
type AIAssistantSharedExpense struct {
	Label    string `json:"label"`
	Currency string `json:"currency"`
	Balance  int64  `json:"balance"`
}
