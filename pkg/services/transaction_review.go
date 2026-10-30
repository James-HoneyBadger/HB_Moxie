package services

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/errs"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/utils"
)

const secondsPerDay = int64(24 * 60 * 60)
const transactionReviewSecondsPerMonth = int64(30 * 24 * 60 * 60)

// TransactionReviewRepository is the data-access boundary used by the
// transaction review service. Keeping it an interface lets the detection
// heuristics be tested without a database.
type TransactionReviewRepository interface {
	GetTransactionsForReview(c core.Context, uid int64, minUnixTime int64) ([]*models.Transaction, error)
}

// TransactionReviewService detects duplicate charges, unusual spending and
// recurring subscriptions from a user's transaction history.
type TransactionReviewService struct {
	repository TransactionReviewRepository
	dismissals TransactionReviewDismissalStore
}

// NewTransactionReviewService returns a review service backed by the given repository.
func NewTransactionReviewService(repository TransactionReviewRepository) *TransactionReviewService {
	return &TransactionReviewService{
		repository: repository,
	}
}

// TransactionReviewDismissalStore persists the findings a user does not want to
// see again. It is separate from the transaction repository so the detection
// heuristics stay testable without a database.
type TransactionReviewDismissalStore interface {
	GetDismissalKeys(c core.Context, uid int64) (map[string]bool, error)
	AddDismissal(c core.Context, uid int64, kind string, key string) error
	RemoveDismissal(c core.Context, uid int64, kind string, key string) error
}

// TransactionReviews is the application-wide transaction review service.
var TransactionReviews = &TransactionReviewService{
	repository: Transactions,
	dismissals: TransactionReviewDismissals,
}

// Review analyses the transactions of the given user over the last months.
func (s *TransactionReviewService) Review(c core.Context, uid int64, request *models.TransactionReviewRequest) (*models.TransactionReviewResponse, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if request == nil {
		request = &models.TransactionReviewRequest{}
	}

	months := request.Months

	if months <= 0 {
		months = models.TransactionReviewDefaultMonths
	} else if months > models.TransactionReviewMaxMonths {
		months = models.TransactionReviewMaxMonths
	}

	options := models.TransactionReviewDefaultOptions()

	if request.ZScoreThreshold > 0 && request.ZScoreThreshold <= 10 {
		options.ZScoreThreshold = request.ZScoreThreshold
	}

	if request.DuplicateMaxBucket > 0 && request.DuplicateMaxBucket <= 20 {
		options.DuplicateMaxBucket = request.DuplicateMaxBucket
	}

	endUnixTime := time.Now().Unix()
	startUnixTime := endUnixTime - int64(months)*transactionReviewSecondsPerMonth

	transactions, err := s.repository.GetTransactionsForReview(c, uid, startUnixTime)

	if err != nil {
		return nil, err
	}

	dismissedKeys := map[string]bool{}

	if s.dismissals != nil {
		keys, err := s.dismissals.GetDismissalKeys(c, uid)

		if err != nil {
			return nil, err
		}

		dismissedKeys = keys
	}

	return buildTransactionReview(transactions, startUnixTime, endUnixTime, options, dismissedKeys), nil
}

// Dismiss hides a finding until it is restored.
func (s *TransactionReviewService) Dismiss(c core.Context, uid int64, request *models.TransactionReviewDismissRequest) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if request == nil || s.dismissals == nil {
		return errs.ErrParameterInvalid
	}

	if !isValidTransactionReviewDismissalKind(request.Kind) {
		return errs.ErrTransactionReviewDismissalInvalid
	}

	key := strings.TrimSpace(request.Key)

	if key == "" {
		return errs.ErrTransactionReviewDismissalInvalid
	}

	return s.dismissals.AddDismissal(c, uid, request.Kind, key)
}

// Restore shows a dismissed finding again.
func (s *TransactionReviewService) Restore(c core.Context, uid int64, request *models.TransactionReviewDismissRequest) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if request == nil || s.dismissals == nil {
		return errs.ErrParameterInvalid
	}

	if !isValidTransactionReviewDismissalKind(request.Kind) {
		return errs.ErrTransactionReviewDismissalInvalid
	}

	return s.dismissals.RemoveDismissal(c, uid, request.Kind, strings.TrimSpace(request.Key))
}

func isValidTransactionReviewDismissalKind(kind string) bool {
	return kind == models.TransactionReviewDismissalKindDuplicate ||
		kind == models.TransactionReviewDismissalKindUnusual ||
		kind == models.TransactionReviewDismissalKindSubscription
}

func buildTransactionReview(transactions []*models.Transaction, startUnixTime int64, endUnixTime int64, options models.TransactionReviewOptions, dismissedKeys map[string]bool) *models.TransactionReviewResponse {
	duplicates := detectDuplicateTransactions(transactions, options)
	unusual := detectUnusualTransactions(transactions, options)
	subscriptions := detectSubscriptions(transactions)

	// Findings the user dismissed are dropped here so every caller sees the same
	// filtered view.
	keptDuplicates := make([]*models.TransactionReviewDuplicateGroup, 0, len(duplicates))

	for _, group := range duplicates {
		if !dismissedKeys[models.TransactionReviewDismissalKindDuplicate+"\x00"+group.Key] {
			keptDuplicates = append(keptDuplicates, group)
		}
	}

	keptUnusual := make([]*models.TransactionReviewUnusual, 0, len(unusual))

	for _, item := range unusual {
		if !dismissedKeys[models.TransactionReviewDismissalKindUnusual+"\x00"+item.Key] {
			keptUnusual = append(keptUnusual, item)
		}
	}

	keptSubscriptions := make([]*models.TransactionReviewSubscription, 0, len(subscriptions))

	for _, subscription := range subscriptions {
		if !dismissedKeys[models.TransactionReviewDismissalKindSubscription+"\x00"+subscription.Key] {
			keptSubscriptions = append(keptSubscriptions, subscription)
		}
	}

	return &models.TransactionReviewResponse{
		StartTime:          startUnixTime,
		EndTime:            endUnixTime,
		TotalTransactions:  int32(len(transactions)),
		ZScoreThreshold:    options.ZScoreThreshold,
		DuplicateMaxBucket: options.DuplicateMaxBucket,
		Duplicates:         keptDuplicates,
		Unusual:            keptUnusual,
		Subscriptions:      keptSubscriptions,
	}
}

func transactionReviewDuplicateKey(accountId int64, txnType models.TransactionDbType, amount int64) string {
	return fmt.Sprintf("%d:%d:%d", accountId, txnType, amount)
}

func transactionReviewUnusualKey(transactionId int64) string {
	return fmt.Sprintf("%d", transactionId)
}

func transactionReviewSubscriptionKey(accountId int64, name string) string {
	return fmt.Sprintf("%d:%s", accountId, strings.ToLower(strings.TrimSpace(name)))
}

// detectDuplicateTransactions finds transactions with the same account, type
// and amount within a short window, which usually means a charge was recorded
// twice or billed twice.
func detectDuplicateTransactions(transactions []*models.Transaction, options models.TransactionReviewOptions) []*models.TransactionReviewDuplicateGroup {
	type bucketKey struct {
		accountId int64
		txnType   models.TransactionDbType
		amount    int64
	}

	windowSeconds := int64(models.TransactionReviewDuplicateWindowDays) * secondsPerDay
	buckets := make(map[bucketKey][]*models.Transaction)

	for _, transaction := range transactions {
		// Transfers move between two accounts, so equal amounts there are expected.
		if transaction.Type != models.TRANSACTION_DB_TYPE_EXPENSE && transaction.Type != models.TRANSACTION_DB_TYPE_INCOME {
			continue
		}

		if transaction.Amount == 0 {
			continue
		}

		key := bucketKey{transaction.AccountId, transaction.Type, transaction.Amount}
		buckets[key] = append(buckets[key], transaction)
	}

	groups := make([]*models.TransactionReviewDuplicateGroup, 0, len(buckets))

	for key, items := range buckets {
		if len(items) < 2 || int32(len(items)) > options.DuplicateMaxBucket {
			continue
		}

		sort.Slice(items, func(i, j int) bool {
			return items[i].TransactionTime < items[j].TransactionTime
		})

		for i := 0; i < len(items); {
			windowEndUnixTime := utils.GetUnixTimeFromTransactionTime(items[i].TransactionTime) + windowSeconds

			j := i + 1

			for j < len(items) && utils.GetUnixTimeFromTransactionTime(items[j].TransactionTime) <= windowEndUnixTime {
				j++
			}

			if j-i >= 2 {
				group := &models.TransactionReviewDuplicateGroup{
					Key:          transactionReviewDuplicateKey(key.accountId, key.txnType, key.amount),
					AccountId:    key.accountId,
					Type:         key.txnType,
					Amount:       key.amount,
					Transactions: make([]*models.TransactionReviewTransaction, 0, j-i),
				}

				for _, item := range items[i:j] {
					group.Transactions = append(group.Transactions, newReviewTransaction(item))
				}

				groups = append(groups, group)
			}

			i = j
		}
	}

	sort.SliceStable(groups, func(i, j int) bool {
		return groupLatestUnixTime(groups[i]) > groupLatestUnixTime(groups[j])
	})

	if len(groups) > models.TransactionReviewDuplicateLimit {
		groups = groups[:models.TransactionReviewDuplicateLimit]
	}

	return groups
}

// detectUnusualTransactions flags expenses far above the usual amount for their
// category, using a z-score over the category's own history.
func detectUnusualTransactions(transactions []*models.Transaction, options models.TransactionReviewOptions) []*models.TransactionReviewUnusual {
	type categoryStats struct {
		mean   float64
		stddev float64
		count  int32
	}

	amountsByCategory := make(map[int64][]float64)

	for _, transaction := range transactions {
		if transaction.Type != models.TRANSACTION_DB_TYPE_EXPENSE || transaction.Amount <= 0 {
			continue
		}

		amountsByCategory[transaction.CategoryId] = append(amountsByCategory[transaction.CategoryId], float64(transaction.Amount))
	}

	statsByCategory := make(map[int64]categoryStats, len(amountsByCategory))

	for categoryId, amounts := range amountsByCategory {
		if len(amounts) < models.TransactionReviewUnusualMinimumSamples {
			continue
		}

		mean, stddev := meanAndStdDev(amounts)

		if stddev <= 0 {
			continue
		}

		statsByCategory[categoryId] = categoryStats{mean: mean, stddev: stddev, count: int32(len(amounts))}
	}

	unusual := make([]*models.TransactionReviewUnusual, 0)

	for _, transaction := range transactions {
		if transaction.Type != models.TRANSACTION_DB_TYPE_EXPENSE || transaction.Amount <= 0 {
			continue
		}

		stats, exists := statsByCategory[transaction.CategoryId]

		if !exists {
			continue
		}

		// Require a meaningful jump as well as a large z-score, so that
		// categories with tiny amounts do not produce noise.
		if float64(transaction.Amount) < stats.mean*1.5 {
			continue
		}

		zScore := (float64(transaction.Amount) - stats.mean) / stats.stddev

		if zScore < options.ZScoreThreshold {
			continue
		}

		unusual = append(unusual, &models.TransactionReviewUnusual{
			Key:            transactionReviewUnusualKey(transaction.TransactionId),
			Transaction:    newReviewTransaction(transaction),
			CategoryMean:   math.Round(stats.mean*100) / 100,
			CategoryStdDev: math.Round(stats.stddev*100) / 100,
			ZScore:         math.Round(zScore*100) / 100,
			SampleCount:    stats.count,
		})
	}

	sort.SliceStable(unusual, func(i, j int) bool {
		if unusual[i].ZScore != unusual[j].ZScore {
			return unusual[i].ZScore > unusual[j].ZScore
		}

		return unusual[i].Transaction.Time > unusual[j].Transaction.Time
	})

	if len(unusual) > models.TransactionReviewUnusualLimit {
		unusual = unusual[:models.TransactionReviewUnusualLimit]
	}

	return unusual
}

// detectSubscriptions groups expenses by normalized description and reports the
// groups that repeat on a regular cadence.
func detectSubscriptions(transactions []*models.Transaction) []*models.TransactionReviewSubscription {
	byMerchant := make(map[string][]*models.Transaction)

	for _, transaction := range transactions {
		if transaction.Type != models.TRANSACTION_DB_TYPE_EXPENSE || transaction.Amount <= 0 {
			continue
		}

		key := normalizeMerchantName(transaction.Comment)

		if key == "" {
			continue
		}

		byMerchant[key] = append(byMerchant[key], transaction)
	}

	subscriptions := make([]*models.TransactionReviewSubscription, 0, len(byMerchant))

	for _, items := range byMerchant {
		if len(items) < models.TransactionReviewSubscriptionMinimumOccurrences {
			continue
		}

		sort.Slice(items, func(i, j int) bool {
			return items[i].TransactionTime < items[j].TransactionTime
		})

		intervals := make([]float64, 0, len(items)-1)

		for i := 1; i < len(items); i++ {
			days := float64(utils.GetUnixTimeFromTransactionTime(items[i].TransactionTime)-utils.GetUnixTimeFromTransactionTime(items[i-1].TransactionTime)) / float64(secondsPerDay)

			if days <= 0 {
				continue
			}

			intervals = append(intervals, days)
		}

		if len(intervals) < 2 {
			continue
		}

		medianDays := median(intervals)
		cadence, exists := classifyCadence(medianDays)

		if !exists {
			continue
		}

		// Reject clusters that merely share a description but happen irregularly.
		consistent := 0

		for _, days := range intervals {
			if math.Abs(days-medianDays) <= medianDays*models.TransactionReviewSubscriptionMaxIntervalDrift {
				consistent++
			}
		}

		if float64(consistent)/float64(len(intervals)) < 0.6 {
			continue
		}

		first := items[0]
		last := items[len(items)-1]

		var totalAmount int64

		for _, item := range items {
			totalAmount += item.Amount
		}

		subscriptions = append(subscriptions, &models.TransactionReviewSubscription{
			Key:           transactionReviewSubscriptionKey(last.AccountId, last.Comment),
			Name:          strings.TrimSpace(last.Comment),
			Cadence:       cadence,
			Occurrences:   int32(len(items)),
			AverageAmount: totalAmount / int64(len(items)),
			FirstAmount:   first.Amount,
			LastAmount:    last.Amount,
			PriceChange:   last.Amount - first.Amount,
			FirstTime:     utils.GetUnixTimeFromTransactionTime(first.TransactionTime),
			LastTime:      utils.GetUnixTimeFromTransactionTime(last.TransactionTime),
			AccountId:     last.AccountId,
			CategoryId:    last.CategoryId,
		})
	}

	sort.SliceStable(subscriptions, func(i, j int) bool {
		if subscriptions[i].PriceChange != subscriptions[j].PriceChange {
			// Largest increases first, so price hikes surface at the top.
			return subscriptions[i].PriceChange > subscriptions[j].PriceChange
		}

		return subscriptions[i].LastTime > subscriptions[j].LastTime
	})

	if len(subscriptions) > models.TransactionReviewSubscriptionLimit {
		subscriptions = subscriptions[:models.TransactionReviewSubscriptionLimit]
	}

	return subscriptions
}

// normalizeMerchantName reduces a free-text comment to a stable merchant key:
// lower case, letters and digits only, numeric-only tokens dropped.
func normalizeMerchantName(comment string) string {
	var builder strings.Builder
	lastWasSeparator := true

	for _, r := range strings.ToLower(comment) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			lastWasSeparator = false
			continue
		}

		if !lastWasSeparator {
			builder.WriteRune(' ')
			lastWasSeparator = true
		}
	}

	tokens := strings.Fields(builder.String())
	kept := make([]string, 0, len(tokens))

	for _, token := range tokens {
		hasLetter := false

		for _, r := range token {
			if unicode.IsLetter(r) {
				hasLetter = true
				break
			}
		}

		if !hasLetter {
			// Reference numbers and card suffixes change every month.
			continue
		}

		kept = append(kept, token)
	}

	if len(kept) > 4 {
		kept = kept[:4]
	}

	return strings.Join(kept, " ")
}

func classifyCadence(medianDays float64) (string, bool) {
	switch {
	case medianDays >= 5 && medianDays <= 9:
		return models.TransactionReviewCadenceWeekly, true
	case medianDays >= 25 && medianDays <= 35:
		return models.TransactionReviewCadenceMonthly, true
	case medianDays >= 85 && medianDays <= 95:
		return models.TransactionReviewCadenceQuarterly, true
	case medianDays >= 350 && medianDays <= 380:
		return models.TransactionReviewCadenceYearly, true
	default:
		return "", false
	}
}

func meanAndStdDev(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}

	var sum float64

	for _, value := range values {
		sum += value
	}

	mean := sum / float64(len(values))

	var squaredSum float64

	for _, value := range values {
		diff := value - mean
		squaredSum += diff * diff
	}

	return mean, math.Sqrt(squaredSum / float64(len(values)))
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	middle := len(sorted) / 2

	if len(sorted)%2 == 1 {
		return sorted[middle]
	}

	return (sorted[middle-1] + sorted[middle]) / 2
}

func newReviewTransaction(transaction *models.Transaction) *models.TransactionReviewTransaction {
	return &models.TransactionReviewTransaction{
		Id:         transaction.TransactionId,
		Type:       transaction.Type,
		Time:       utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime),
		Amount:     transaction.Amount,
		Comment:    transaction.Comment,
		CategoryId: transaction.CategoryId,
		AccountId:  transaction.AccountId,
	}
}

func groupLatestUnixTime(group *models.TransactionReviewDuplicateGroup) int64 {
	var latest int64

	for _, transaction := range group.Transactions {
		if transaction.Time > latest {
			latest = transaction.Time
		}
	}

	return latest
}
