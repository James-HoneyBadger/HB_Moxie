package errs

import "net/http"

// Error codes related to transaction splits (shared expenses)
var (
	ErrTransactionSplitPersonNameInvalid   = NewNormalError(NormalSubcategoryTransactionSplit, 0, http.StatusBadRequest, "split person name is invalid")
	ErrTransactionSplitAmountInvalid       = NewNormalError(NormalSubcategoryTransactionSplit, 1, http.StatusBadRequest, "split amount is invalid")
	ErrTransactionSplitPersonDuplicated    = NewNormalError(NormalSubcategoryTransactionSplit, 2, http.StatusBadRequest, "split person is duplicated")
	ErrTransactionSplitTooManyPeople       = NewNormalError(NormalSubcategoryTransactionSplit, 3, http.StatusBadRequest, "too many people to split with")
	ErrTransactionSplitTotalAmountExceeded = NewNormalError(NormalSubcategoryTransactionSplit, 4, http.StatusBadRequest, "total split amount exceeds transaction amount")
	ErrTransactionSplitNotSupported        = NewNormalError(NormalSubcategoryTransactionSplit, 5, http.StatusBadRequest, "only expense transactions can be split")
)
