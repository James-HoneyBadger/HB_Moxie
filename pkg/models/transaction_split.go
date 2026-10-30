package models

// TransactionSplitMaxCount is the maximum number of people a transaction can be split with
const TransactionSplitMaxCount = 20

// TransactionSplitMaxPersonNameLength is the maximum length of a split person name
const TransactionSplitMaxPersonNameLength = 64

// TransactionSplit represents a share of a transaction owed by another person
type TransactionSplit struct {
	SplitId       int64  `xorm:"PK"`
	Uid           int64  `xorm:"INDEX(IDX_transaction_split_uid_transaction_id) INDEX(IDX_transaction_split_uid_person_name) NOT NULL"`
	TransactionId int64  `xorm:"INDEX(IDX_transaction_split_uid_transaction_id) NOT NULL"`
	PersonName    string `xorm:"VARCHAR(64) NOT NULL"`
	// Currency is captured when the split is created so settlement totals never
	// mix currencies.
	Currency        string `xorm:"VARCHAR(3) NOT NULL"`
	Amount          int64  `xorm:"NOT NULL"`
	CreatedUnixTime int64  `xorm:"NOT NULL"`
}

// TransactionSplitRequestItem represents one person's share in a split update request
type TransactionSplitRequestItem struct {
	PersonName string `json:"personName"`
	Amount     int64  `json:"amount"`
}

// TransactionSplitUpdateRequest represents all parameters of a split update request
type TransactionSplitUpdateRequest struct {
	TransactionId int64                         `json:"transactionId,string" binding:"required,min=1"`
	Splits        []TransactionSplitRequestItem `json:"splits"`
}

// TransactionSplitInfoResponse represents one person's share of a transaction
type TransactionSplitInfoResponse struct {
	PersonName string `json:"personName"`
	Currency   string `json:"currency"`
	Amount     int64  `json:"amount"`
}

// TransactionSplitsResponse represents the splits of one transaction
type TransactionSplitsResponse struct {
	TransactionId     int64                           `json:"transactionId,string"`
	TransactionAmount int64                           `json:"transactionAmount"`
	Currency          string                          `json:"currency"`
	Splits            []*TransactionSplitInfoResponse `json:"splits"`
}

// SharedExpenseSummaryItem represents what one person owes the user in one currency
type SharedExpenseSummaryItem struct {
	PersonName string `json:"personName"`
	Currency   string `json:"currency"`
	Total      int64  `json:"total"`
	EntryCount int32  `json:"entryCount"`
	LastTime   int64  `json:"lastTime"`
}

// SharedExpenseEntryResponse represents one shared expense of a person
type SharedExpenseEntryResponse struct {
	TransactionId     int64  `json:"transactionId,string"`
	Time              int64  `json:"time"`
	UtcOffset         int16  `json:"utcOffset"`
	CategoryId        int64  `json:"categoryId,string"`
	AccountId         int64  `json:"accountId,string"`
	TransactionAmount int64  `json:"transactionAmount"`
	SplitAmount       int64  `json:"splitAmount"`
	Currency          string `json:"currency"`
	Comment           string `json:"comment"`
}
