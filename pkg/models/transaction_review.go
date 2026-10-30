package models

// Transaction review parameters
const (
	TransactionReviewDefaultMonths = 12
	TransactionReviewMinMonths     = 1
	TransactionReviewMaxMonths     = 60

	TransactionReviewDuplicateWindowDays = 3
	TransactionReviewDuplicateLimit      = 20
	// The same amount recurring more often than this on one account is a habit
	// (a daily commute, a weekly shop), not a duplicated charge.
	TransactionReviewDuplicateMaxBucketSize = 4

	TransactionReviewUnusualMinimumSamples  = 8
	TransactionReviewUnusualZScoreThreshold = 3.0
	TransactionReviewUnusualLimit           = 20

	TransactionReviewSubscriptionMinimumOccurrences = 3
	TransactionReviewSubscriptionMaxIntervalDrift   = 0.4
	TransactionReviewSubscriptionLimit              = 50
)

// Transaction review dismissal kinds
const (
	TransactionReviewDismissalKindDuplicate    = "duplicate"
	TransactionReviewDismissalKindUnusual      = "unusual"
	TransactionReviewDismissalKindSubscription = "subscription"
)

// TransactionReviewRequest represents the parameters of a transaction review request.
// The two thresholds are optional and fall back to the defaults above, so the same
// review can be tuned per user without a settings migration.
type TransactionReviewRequest struct {
	Months             int32   `form:"months"`
	ZScoreThreshold    float64 `form:"z_score"`
	DuplicateMaxBucket int32   `form:"duplicate_max_bucket"`
}

// TransactionReviewOptions are the thresholds a review runs with
type TransactionReviewOptions struct {
	ZScoreThreshold    float64
	DuplicateMaxBucket int32
}

// TransactionReviewDefaultOptions returns the thresholds a review uses when the
// request does not override them
func TransactionReviewDefaultOptions() TransactionReviewOptions {
	return TransactionReviewOptions{
		ZScoreThreshold:    TransactionReviewUnusualZScoreThreshold,
		DuplicateMaxBucket: TransactionReviewDuplicateMaxBucketSize,
	}
}

// TransactionReviewDismissal records that the user does not want to see a finding again
type TransactionReviewDismissal struct {
	DismissalId     int64  `xorm:"PK"`
	Uid             int64  `xorm:"INDEX(IDX_transaction_review_dismissal_uid_kind_key) NOT NULL"`
	Kind            string `xorm:"VARCHAR(16) NOT NULL"`
	Key             string `xorm:"VARCHAR(128) NOT NULL"`
	CreatedUnixTime int64  `xorm:"NOT NULL"`
}

// TransactionReviewDismissRequest represents all parameters of a dismissal request
type TransactionReviewDismissRequest struct {
	Kind string `json:"kind" binding:"required"`
	Key  string `json:"key" binding:"required,max=128"`
}

// TransactionReviewTransaction is a compact transaction reference used by review findings
type TransactionReviewTransaction struct {
	Id         int64             `json:"id,string"`
	Type       TransactionDbType `json:"type"`
	Time       int64             `json:"time"`
	Amount     int64             `json:"amount"`
	Comment    string            `json:"comment"`
	CategoryId int64             `json:"categoryId,string"`
	AccountId  int64             `json:"accountId,string"`
}

// TransactionReviewDuplicateGroup represents transactions with the same amount, account and type close together in time
type TransactionReviewDuplicateGroup struct {
	// Key identifies the finding when dismissing it
	Key          string                          `json:"key"`
	AccountId    int64                           `json:"accountId,string"`
	Type         TransactionDbType               `json:"type"`
	Amount       int64                           `json:"amount"`
	Transactions []*TransactionReviewTransaction `json:"transactions"`
}

// TransactionReviewUnusual represents a transaction far above the usual amount for its category
type TransactionReviewUnusual struct {
	// Key identifies the finding when dismissing it
	Key            string                        `json:"key"`
	Transaction    *TransactionReviewTransaction `json:"transaction"`
	CategoryMean   float64                       `json:"categoryMean"`
	CategoryStdDev float64                       `json:"categoryStdDev"`
	ZScore         float64                       `json:"zScore"`
	SampleCount    int32                         `json:"sampleCount"`
}

// Transaction review cadence values
const (
	TransactionReviewCadenceWeekly    = "weekly"
	TransactionReviewCadenceMonthly   = "monthly"
	TransactionReviewCadenceQuarterly = "quarterly"
	TransactionReviewCadenceYearly    = "yearly"
)

// TransactionReviewSubscription represents a recurring expense detected from transaction history
type TransactionReviewSubscription struct {
	// Key identifies the finding when dismissing it
	Key           string `json:"key"`
	Name          string `json:"name"`
	Cadence       string `json:"cadence"`
	Occurrences   int32  `json:"occurrences"`
	AverageAmount int64  `json:"averageAmount"`
	FirstAmount   int64  `json:"firstAmount"`
	LastAmount    int64  `json:"lastAmount"`
	PriceChange   int64  `json:"priceChange"`
	FirstTime     int64  `json:"firstTime"`
	LastTime      int64  `json:"lastTime"`
	AccountId     int64  `json:"accountId,string"`
	CategoryId    int64  `json:"categoryId,string"`
}

// TransactionReviewResponse is the result of a transaction review
type TransactionReviewResponse struct {
	StartTime          int64                              `json:"startTime"`
	EndTime            int64                              `json:"endTime"`
	TotalTransactions  int32                              `json:"totalTransactions"`
	ZScoreThreshold    float64                            `json:"zScoreThreshold"`
	DuplicateMaxBucket int32                              `json:"duplicateMaxBucket"`
	Duplicates         []*TransactionReviewDuplicateGroup `json:"duplicates"`
	Unusual            []*TransactionReviewUnusual        `json:"unusual"`
	Subscriptions      []*TransactionReviewSubscription   `json:"subscriptions"`
}
