package models

// TransactionSplitMaxCount is the maximum number of people a transaction can be split with
const TransactionSplitMaxCount = 20

// TransactionSplitMaxPersonNameLength is the maximum length of a split person name
const TransactionSplitMaxPersonNameLength = 64

// TransactionSplitDirection represents who owes whom for a share of a transaction
type TransactionSplitDirection int8

// Split directions
const (
	// SPLIT_DIRECTION_THEY_OWE_ME means the other person owes the user their share
	SPLIT_DIRECTION_THEY_OWE_ME TransactionSplitDirection = 1
	// SPLIT_DIRECTION_I_OWE_THEM means the user owes the other person their share
	SPLIT_DIRECTION_I_OWE_THEM TransactionSplitDirection = 2
)

// Values of the shared_filter transaction list parameter
const (
	// TRANSACTION_SHARED_FILTER_SHARED keeps only transactions that are split
	TRANSACTION_SHARED_FILTER_SHARED = "shared"
	// TRANSACTION_SHARED_FILTER_UNSHARED keeps only transactions that are not split
	TRANSACTION_SHARED_FILTER_UNSHARED = "unshared"
)

// SettlementMaxAmount is the largest amount a single settlement may record
const SettlementMaxAmount = 99999999999

// TransactionSplit represents a share of a transaction owed by another person, or
// owed to another person when the user did not pay
type TransactionSplit struct {
	SplitId       int64  `xorm:"PK"`
	Uid           int64  `xorm:"INDEX(IDX_transaction_split_uid_transaction_id) INDEX(IDX_transaction_split_uid_person_name) NOT NULL"`
	TransactionId int64  `xorm:"INDEX(IDX_transaction_split_uid_transaction_id) NOT NULL"`
	PersonName    string `xorm:"VARCHAR(64) NOT NULL"`
	// Currency is captured when the split is created so settlement totals never
	// mix currencies.
	Currency        string                    `xorm:"VARCHAR(3) NOT NULL"`
	Amount          int64                     `xorm:"NOT NULL"`
	Direction       TransactionSplitDirection `xorm:"NOT NULL"`
	CreatedUnixTime int64                     `xorm:"NOT NULL"`
}

// SharedExpenseSettlement represents a payment between the user and another person
// that reduces what they owe each other
type SharedExpenseSettlement struct {
	SettlementId int64  `xorm:"PK"`
	Uid          int64  `xorm:"INDEX(IDX_shared_expense_settlement_uid_person_name) NOT NULL"`
	PersonName   string `xorm:"VARCHAR(64) NOT NULL"`
	Currency     string `xorm:"VARCHAR(3) NOT NULL"`
	// Amount is positive when the other person pays the user and negative when the
	// user pays the other person.
	Amount          int64  `xorm:"NOT NULL"`
	Time            int64  `xorm:"NOT NULL"`
	Comment         string `xorm:"VARCHAR(255) NOT NULL"`
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
	Direction     TransactionSplitDirection     `json:"direction"`
	Splits        []TransactionSplitRequestItem `json:"splits"`
}

// TransactionSplitInfoResponse represents one person's share of a transaction
type TransactionSplitInfoResponse struct {
	PersonName string `json:"personName"`
	Currency   string `json:"currency"`
	Amount     int64  `json:"amount"`
}

// TransactionSplitRenamePersonRequest represents all parameters of a person rename request
type TransactionSplitRenamePersonRequest struct {
	OldName string `json:"oldName" binding:"required"`
	NewName string `json:"newName" binding:"required"`
}

// SharedExpenseSettlementRequest represents all parameters of a settlement request
type SharedExpenseSettlementRequest struct {
	PersonName string `json:"personName" binding:"required"`
	Currency   string `json:"currency" binding:"required,len=3"`
	Amount     int64  `json:"amount"`
	Time       int64  `json:"time" binding:"min=0"`
	Comment    string `json:"comment" binding:"max=255"`
}

// SharedExpenseSettlementDeleteRequest represents all parameters of a settlement deletion request
type SharedExpenseSettlementDeleteRequest struct {
	SettlementId int64 `json:"settlementId,string" binding:"required,min=1"`
}

// SharedExpenseSettlementInfoResponse represents one recorded settlement
type SharedExpenseSettlementInfoResponse struct {
	SettlementId int64  `json:"settlementId,string"`
	PersonName   string `json:"personName"`
	Currency     string `json:"currency"`
	Amount       int64  `json:"amount"`
	Time         int64  `json:"time"`
	Comment      string `json:"comment"`
}

// TransactionSplitsResponse represents the splits of one transaction
type TransactionSplitsResponse struct {
	TransactionId     int64                           `json:"transactionId,string"`
	TransactionAmount int64                           `json:"transactionAmount"`
	Currency          string                          `json:"currency"`
	Direction         TransactionSplitDirection       `json:"direction"`
	Splits            []*TransactionSplitInfoResponse `json:"splits"`
}

// SharedExpenseSummaryItem represents the balance between the user and one person in
// one currency
type SharedExpenseSummaryItem struct {
	PersonName string `json:"personName"`
	Currency   string `json:"currency"`
	// Total is the outstanding balance: positive when the person owes the user,
	// negative when the user owes the person.
	Total int64 `json:"total"`
	// OwedToMe and IOwe report the gross shared expenses before settlements.
	OwedToMe   int64 `json:"owedToMe"`
	IOwe       int64 `json:"iOwe"`
	Settled    int64 `json:"settled"`
	EntryCount int32 `json:"entryCount"`
	LastTime   int64 `json:"lastTime"`
}

// SharedExpenseEntryResponse represents one shared expense or settlement of a person
type SharedExpenseEntryResponse struct {
	TransactionId     int64 `json:"transactionId,string"`
	Time              int64 `json:"time"`
	UtcOffset         int16 `json:"utcOffset"`
	CategoryId        int64 `json:"categoryId,string"`
	AccountId         int64 `json:"accountId,string"`
	TransactionAmount int64 `json:"transactionAmount"`
	// SplitAmount is positive when the person owes the user and negative when the
	// user owes the person.
	SplitAmount  int64                     `json:"splitAmount"`
	Currency     string                    `json:"currency"`
	Comment      string                    `json:"comment"`
	Direction    TransactionSplitDirection `json:"direction"`
	IsSettlement bool                      `json:"isSettlement"`
	SettlementId int64                     `json:"settlementId,string"`
}

// TransactionSplitIndexItem represents the split state of one transaction, used to
// mark split transactions in the transaction list
type TransactionSplitIndexItem struct {
	TransactionId int64  `json:"transactionId,string"`
	Total         int64  `json:"total"`
	Currency      string `json:"currency"`
	PersonCount   int32  `json:"personCount"`
}
