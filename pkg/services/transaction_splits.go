package services

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/datastore"
	"github.com/James-HoneyBadger/honey-ledger/pkg/errs"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/utils"
	"github.com/James-HoneyBadger/honey-ledger/pkg/uuid"

	"xorm.io/xorm"
)

// TransactionSplitService records how much of a transaction other people owe the
// user, and summarizes what each person owes in total.
type TransactionSplitService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a transaction split service singleton instance
var (
	TransactionSplits = &TransactionSplitService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetTransactionSplits returns the splits recorded for one transaction.
func (s *TransactionSplitService) GetTransactionSplits(c core.Context, uid int64, transactionId int64) (*models.TransactionSplitsResponse, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if transactionId <= 0 {
		return nil, errs.ErrTransactionNotFound
	}

	session := s.UserDataDB(uid).NewSession(c)
	defer session.Close()

	var transaction models.Transaction

	exists, err := session.Where("uid=? AND deleted=? AND transaction_id=?", uid, false, transactionId).Get(&transaction)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errs.ErrTransactionNotFound
	}

	var splits []*models.TransactionSplit

	if err := session.Where("uid=? AND transaction_id=?", uid, transactionId).OrderBy("created_unix_time asc").Find(&splits); err != nil {
		return nil, err
	}

	response := &models.TransactionSplitsResponse{
		TransactionId:     transactionId,
		TransactionAmount: transaction.Amount,
		Currency:          s.getAccountCurrency(session, transaction.AccountId),
		Direction:         models.SPLIT_DIRECTION_THEY_OWE_ME,
		Splits:            make([]*models.TransactionSplitInfoResponse, 0, len(splits)),
	}

	if len(splits) > 0 {
		response.Direction = splits[0].Direction
	}

	for _, split := range splits {
		response.Splits = append(response.Splits, &models.TransactionSplitInfoResponse{
			PersonName: split.PersonName,
			Currency:   split.Currency,
			Amount:     split.Amount,
		})
	}

	return response, nil
}

// SetTransactionSplits replaces the splits of one transaction.
func (s *TransactionSplitService) SetTransactionSplits(c core.Context, uid int64, request *models.TransactionSplitUpdateRequest) (*models.TransactionSplitsResponse, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if request == nil || request.TransactionId <= 0 {
		return nil, errs.ErrTransactionNotFound
	}

	direction := request.Direction

	if direction == 0 {
		direction = models.SPLIT_DIRECTION_THEY_OWE_ME
	}

	if direction != models.SPLIT_DIRECTION_THEY_OWE_ME && direction != models.SPLIT_DIRECTION_I_OWE_THEM {
		return nil, errs.ErrTransactionSplitDirectionInvalid
	}

	normalized, err := normalizeTransactionSplits(request.Splits)

	if err != nil {
		return nil, err
	}

	database := s.UserDataDB(uid)
	session := database.NewSession(c)
	defer session.Close()

	var transaction models.Transaction

	exists, err := session.Where("uid=? AND deleted=? AND transaction_id=?", uid, false, request.TransactionId).Get(&transaction)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errs.ErrTransactionNotFound
	}

	if transaction.Type != models.TRANSACTION_DB_TYPE_EXPENSE {
		return nil, errs.ErrTransactionSplitNotSupported
	}

	currency := s.getAccountCurrency(session, transaction.AccountId)

	var total int64

	for _, split := range normalized {
		total += split.Amount
	}

	// Nobody can owe more than the transaction cost.
	if total > int64(math.Abs(float64(transaction.Amount))) {
		return nil, errs.ErrTransactionSplitTotalAmountExceeded
	}

	now := time.Now().Unix()
	splits := make([]*models.TransactionSplit, 0, len(normalized))

	for _, split := range normalized {
		splits = append(splits, &models.TransactionSplit{
			SplitId:         s.GenerateUuid(uuid.UUID_TYPE_USER),
			Uid:             uid,
			TransactionId:   request.TransactionId,
			PersonName:      split.PersonName,
			Currency:        currency,
			Amount:          split.Amount,
			Direction:       direction,
			CreatedUnixTime: now,
		})
	}

	err = database.DoTransaction(c, func(sess *xorm.Session) error {
		if _, err := sess.Where("uid=? AND transaction_id=?", uid, request.TransactionId).Delete(new(models.TransactionSplit)); err != nil {
			return err
		}

		if len(splits) == 0 {
			return nil
		}

		_, err := sess.Insert(splits)
		return err
	})

	if err != nil {
		return nil, err
	}

	response := &models.TransactionSplitsResponse{
		TransactionId:     request.TransactionId,
		TransactionAmount: transaction.Amount,
		Currency:          currency,
		Direction:         direction,
		Splits:            make([]*models.TransactionSplitInfoResponse, 0, len(splits)),
	}

	for _, split := range splits {
		response.Splits = append(response.Splits, &models.TransactionSplitInfoResponse{
			PersonName: split.PersonName,
			Currency:   split.Currency,
			Amount:     split.Amount,
		})
	}

	return response, nil
}

// GetSharedExpenseSummary returns what each person owes the user, grouped by
// currency. Splits of deleted transactions are ignored.
func (s *TransactionSplitService) GetSharedExpenseSummary(c core.Context, uid int64) ([]*models.SharedExpenseSummaryItem, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	session := s.UserDataDB(uid).NewSession(c)
	defer session.Close()

	var splits []*models.TransactionSplit

	if err := session.Where("uid=?", uid).Find(&splits); err != nil {
		return nil, err
	}

	var settlements []*models.SharedExpenseSettlement

	if err := session.Where("uid=?", uid).Find(&settlements); err != nil {
		return nil, err
	}

	if len(splits) == 0 && len(settlements) == 0 {
		return []*models.SharedExpenseSummaryItem{}, nil
	}

	transactionIdSet := make(map[int64]bool, len(splits))
	transactionIds := make([]int64, 0, len(splits))

	for _, split := range splits {
		if transactionIdSet[split.TransactionId] {
			continue
		}

		transactionIdSet[split.TransactionId] = true
		transactionIds = append(transactionIds, split.TransactionId)
	}

	var deletedTransactions []*models.Transaction

	if err := session.Where("uid=? AND deleted=?", uid, true).In("transaction_id", transactionIds).Cols("transaction_id").Find(&deletedTransactions); err != nil {
		return nil, err
	}

	deletedTransactionIds := make(map[int64]bool, len(deletedTransactions))

	for _, transaction := range deletedTransactions {
		deletedTransactionIds[transaction.TransactionId] = true
	}

	return summarizeSharedExpenses(splits, settlements, deletedTransactionIds), nil
}

// GetTransactionSplitIndex returns the split state of the transactions in a time
// range, so the transaction list can mark the ones that are shared. Times are in
// the same millisecond scale as the transaction list filters; a zero maxTime
// leaves the range open ended.
func (s *TransactionSplitService) GetTransactionSplitIndex(c core.Context, uid int64, minTime int64, maxTime int64) ([]*models.TransactionSplitIndexItem, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	session := s.UserDataDB(uid).NewSession(c)
	defer session.Close()

	var splits []*models.TransactionSplit

	if err := session.Where("uid=?", uid).Find(&splits); err != nil {
		return nil, err
	}

	if len(splits) == 0 {
		return []*models.TransactionSplitIndexItem{}, nil
	}

	transactionIds := make([]int64, 0, len(splits))

	for _, split := range splits {
		transactionIds = append(transactionIds, split.TransactionId)
	}

	var transactions []*models.Transaction
	query := session.Where("uid=? AND deleted=?", uid, false).In("transaction_id", transactionIds)

	if minTime > 0 {
		query = query.And("transaction_time >= ?", minTime)
	}

	if maxTime > 0 {
		query = query.And("transaction_time <= ?", maxTime)
	}

	if err := query.Find(&transactions); err != nil {
		return nil, err
	}

	inRange := make(map[int64]bool, len(transactions))

	for _, transaction := range transactions {
		inRange[transaction.TransactionId] = true
	}

	index := make(map[int64]*models.TransactionSplitIndexItem, len(splits))

	for _, split := range splits {
		if !inRange[split.TransactionId] {
			continue
		}

		item, exists := index[split.TransactionId]

		if !exists {
			item = &models.TransactionSplitIndexItem{
				TransactionId: split.TransactionId,
				Currency:      split.Currency,
			}
			index[split.TransactionId] = item
		}

		item.Total += split.Amount
		item.PersonCount++
	}

	result := make([]*models.TransactionSplitIndexItem, 0, len(index))

	for _, item := range index {
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TransactionId > result[j].TransactionId
	})

	return result, nil
}

// GetSharedExpensesByPerson returns the individual shared expenses owed by one person,
// newest first and excluding deleted transactions.
func (s *TransactionSplitService) GetSharedExpensesByPerson(c core.Context, uid int64, personName string) ([]*models.SharedExpenseEntryResponse, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	personName = strings.TrimSpace(personName)

	if personName == "" {
		return nil, errs.ErrTransactionSplitPersonNameInvalid
	}

	session := s.UserDataDB(uid).NewSession(c)
	defer session.Close()

	var splits []*models.TransactionSplit

	if err := session.Where("uid=? AND person_name=?", uid, personName).Find(&splits); err != nil {
		return nil, err
	}

	var settlements []*models.SharedExpenseSettlement

	if err := session.Where("uid=? AND person_name=?", uid, personName).Find(&settlements); err != nil {
		return nil, err
	}

	if len(splits) == 0 && len(settlements) == 0 {
		return []*models.SharedExpenseEntryResponse{}, nil
	}

	transactionIds := make([]int64, 0, len(splits))

	for _, split := range splits {
		transactionIds = append(transactionIds, split.TransactionId)
	}

	var transactions []*models.Transaction

	if err := session.Where("uid=? AND deleted=?", uid, false).In("transaction_id", transactionIds).Find(&transactions); err != nil {
		return nil, err
	}

	transactionMap := make(map[int64]*models.Transaction, len(transactions))

	for _, transaction := range transactions {
		transactionMap[transaction.TransactionId] = transaction
	}

	entries := make([]*models.SharedExpenseEntryResponse, 0, len(splits))

	for _, split := range splits {
		transaction, exists := transactionMap[split.TransactionId]

		if !exists {
			continue
		}

		splitAmount := split.Amount

		if split.Direction == models.SPLIT_DIRECTION_I_OWE_THEM {
			splitAmount = -splitAmount
		}

		entries = append(entries, &models.SharedExpenseEntryResponse{
			TransactionId:     transaction.TransactionId,
			Time:              utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime),
			UtcOffset:         transaction.TimezoneUtcOffset,
			CategoryId:        transaction.CategoryId,
			AccountId:         transaction.AccountId,
			TransactionAmount: transaction.Amount,
			SplitAmount:       splitAmount,
			Currency:          split.Currency,
			Comment:           transaction.Comment,
			Direction:         split.Direction,
		})
	}

	for _, settlement := range settlements {
		entries = append(entries, &models.SharedExpenseEntryResponse{
			Time:         settlement.Time,
			SplitAmount:  settlement.Amount,
			Currency:     settlement.Currency,
			Comment:      settlement.Comment,
			IsSettlement: true,
			SettlementId: settlement.SettlementId,
		})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Time != entries[j].Time {
			return entries[i].Time > entries[j].Time
		}

		return entries[i].TransactionId > entries[j].TransactionId
	})

	return entries, nil
}

// SettleSharedExpense records a payment between the user and another person. A
// positive amount means the person paid the user, a negative amount means the user
// paid the person.
func (s *TransactionSplitService) SettleSharedExpense(c core.Context, uid int64, request *models.SharedExpenseSettlementRequest) (*models.SharedExpenseSettlementInfoResponse, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if request == nil {
		return nil, errs.ErrParameterInvalid
	}

	personName := strings.TrimSpace(request.PersonName)

	if personName == "" || len([]rune(personName)) > models.TransactionSplitMaxPersonNameLength {
		return nil, errs.ErrTransactionSplitPersonNameInvalid
	}

	if request.Amount == 0 || request.Amount > models.SettlementMaxAmount || request.Amount < -models.SettlementMaxAmount {
		return nil, errs.ErrSharedExpenseSettlementAmountInvalid
	}

	settlementTime := request.Time

	if settlementTime <= 0 {
		settlementTime = time.Now().Unix()
	}

	now := time.Now().Unix()

	settlement := &models.SharedExpenseSettlement{
		SettlementId:    s.GenerateUuid(uuid.UUID_TYPE_USER),
		Uid:             uid,
		PersonName:      personName,
		Currency:        request.Currency,
		Amount:          request.Amount,
		Time:            settlementTime,
		Comment:         request.Comment,
		CreatedUnixTime: now,
	}

	if _, err := s.UserDataDB(uid).NewSession(c).Insert(settlement); err != nil {
		return nil, err
	}

	return &models.SharedExpenseSettlementInfoResponse{
		SettlementId: settlement.SettlementId,
		PersonName:   settlement.PersonName,
		Currency:     settlement.Currency,
		Amount:       settlement.Amount,
		Time:         settlement.Time,
		Comment:      settlement.Comment,
	}, nil
}

// DeleteSharedExpenseSettlement removes a recorded settlement.
func (s *TransactionSplitService) DeleteSharedExpenseSettlement(c core.Context, uid int64, settlementId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if settlementId <= 0 {
		return errs.ErrSharedExpenseSettlementNotFound
	}

	affected, err := s.UserDataDB(uid).NewSession(c).
		Where("uid=? AND settlement_id=?", uid, settlementId).
		Delete(new(models.SharedExpenseSettlement))

	if err != nil {
		return err
	}

	if affected == 0 {
		return errs.ErrSharedExpenseSettlementNotFound
	}

	return nil
}

// GetKnownPeople returns every person the user has shared expenses or settlements
// with, so the interface can suggest names instead of asking for free text.
func (s *TransactionSplitService) GetKnownPeople(c core.Context, uid int64) ([]string, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	session := s.UserDataDB(uid).NewSession(c)
	defer session.Close()

	var splits []*models.TransactionSplit

	if err := session.Where("uid=?", uid).Cols("person_name").Find(&splits); err != nil {
		return nil, err
	}

	var settlements []*models.SharedExpenseSettlement

	if err := session.Where("uid=?", uid).Cols("person_name").Find(&settlements); err != nil {
		return nil, err
	}

	names := make(map[string]bool, len(splits)+len(settlements))

	for _, split := range splits {
		names[split.PersonName] = true
	}

	for _, settlement := range settlements {
		names[settlement.PersonName] = true
	}

	people := make([]string, 0, len(names))

	for name := range names {
		people = append(people, name)
	}

	sort.Strings(people)

	return people, nil
}

// RenameSharedExpensePerson renames a person across every split and settlement, so a
// typo does not leave one person split over several rows. Merging into an existing
// person is rejected rather than silently combining their balances.
func (s *TransactionSplitService) RenameSharedExpensePerson(c core.Context, uid int64, request *models.TransactionSplitRenamePersonRequest) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if request == nil {
		return errs.ErrParameterInvalid
	}

	oldName := strings.TrimSpace(request.OldName)
	newName := strings.TrimSpace(request.NewName)

	if oldName == "" || newName == "" || len([]rune(newName)) > models.TransactionSplitMaxPersonNameLength {
		return errs.ErrTransactionSplitPersonNameInvalid
	}

	if oldName == newName {
		return nil
	}

	database := s.UserDataDB(uid)

	return database.DoTransaction(c, func(sess *xorm.Session) error {
		existing := &models.TransactionSplit{}
		hasExistingSplit, err := sess.Where("uid=? AND person_name=?", uid, newName).Exist(existing)

		if err != nil {
			return err
		}

		existingSettlement := &models.SharedExpenseSettlement{}
		hasExistingSettlement, err := sess.Where("uid=? AND person_name=?", uid, newName).Exist(existingSettlement)

		if err != nil {
			return err
		}

		if hasExistingSplit || hasExistingSettlement {
			return errs.ErrTransactionSplitPersonAlreadyExists
		}

		if _, err := sess.Where("uid=? AND person_name=?", uid, oldName).
			Cols("person_name").
			Update(&models.TransactionSplit{PersonName: newName}); err != nil {
			return err
		}

		_, err = sess.Where("uid=? AND person_name=?", uid, oldName).
			Cols("person_name").
			Update(&models.SharedExpenseSettlement{PersonName: newName})

		return err
	})
}

// normalizeTransactionSplits validates and trims the requested splits. Keeping
// this pure makes the validation rules testable without a database.
func normalizeTransactionSplits(items []models.TransactionSplitRequestItem) ([]models.TransactionSplitRequestItem, error) {
	if len(items) > models.TransactionSplitMaxCount {
		return nil, errs.ErrTransactionSplitTooManyPeople
	}

	normalized := make([]models.TransactionSplitRequestItem, 0, len(items))
	seenNames := make(map[string]bool, len(items))

	for _, item := range items {
		name := strings.TrimSpace(item.PersonName)

		if name == "" || len([]rune(name)) > models.TransactionSplitMaxPersonNameLength {
			return nil, errs.ErrTransactionSplitPersonNameInvalid
		}

		if item.Amount <= 0 {
			return nil, errs.ErrTransactionSplitAmountInvalid
		}

		key := strings.ToLower(name)

		if seenNames[key] {
			return nil, errs.ErrTransactionSplitPersonDuplicated
		}

		seenNames[key] = true
		normalized = append(normalized, models.TransactionSplitRequestItem{PersonName: name, Amount: item.Amount})
	}

	return normalized, nil
}

func summarizeSharedExpenses(splits []*models.TransactionSplit, settlements []*models.SharedExpenseSettlement, deletedTransactionIds map[int64]bool) []*models.SharedExpenseSummaryItem {
	type summaryKey struct {
		personName string
		currency   string
	}

	totals := make(map[summaryKey]*models.SharedExpenseSummaryItem, len(splits))

	getItem := func(personName string, currency string) *models.SharedExpenseSummaryItem {
		key := summaryKey{personName, currency}
		item, exists := totals[key]

		if !exists {
			item = &models.SharedExpenseSummaryItem{
				PersonName: personName,
				Currency:   currency,
			}
			totals[key] = item
		}

		return item
	}

	for _, split := range splits {
		if deletedTransactionIds[split.TransactionId] {
			continue
		}

		item := getItem(split.PersonName, split.Currency)

		if split.Direction == models.SPLIT_DIRECTION_I_OWE_THEM {
			item.IOwe += split.Amount
		} else {
			item.OwedToMe += split.Amount
		}

		item.EntryCount++

		if split.CreatedUnixTime > item.LastTime {
			item.LastTime = split.CreatedUnixTime
		}
	}

	for _, settlement := range settlements {
		item := getItem(settlement.PersonName, settlement.Currency)
		item.Settled += settlement.Amount

		if settlement.CreatedUnixTime > item.LastTime {
			item.LastTime = settlement.CreatedUnixTime
		}
	}

	summary := make([]*models.SharedExpenseSummaryItem, 0, len(totals))

	for _, item := range totals {
		item.Total = item.OwedToMe - item.IOwe - item.Settled
		summary = append(summary, item)
	}

	// People who owe nothing either way no longer need tracking.
	visible := make([]*models.SharedExpenseSummaryItem, 0, len(summary))

	for _, item := range summary {
		if item.Total != 0 {
			visible = append(visible, item)
		}
	}

	sort.SliceStable(visible, func(i, j int) bool {
		if visible[i].Currency != visible[j].Currency {
			return visible[i].Currency < visible[j].Currency
		}

		if (visible[i].Total > 0) != (visible[j].Total > 0) {
			return visible[i].Total > 0
		}

		absI, absJ := visible[i].Total, visible[j].Total

		if absI < 0 {
			absI = -absI
		}

		if absJ < 0 {
			absJ = -absJ
		}

		return absI > absJ
	})

	return visible
}

func (s *TransactionSplitService) getAccountCurrency(session *xorm.Session, accountId int64) string {
	var account models.Account

	has, err := session.ID(accountId).Get(&account)

	if err != nil || !has {
		return ""
	}

	return account.Currency
}
