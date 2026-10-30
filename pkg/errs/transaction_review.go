package errs

import "net/http"

// Error codes related to the transaction review
var (
	ErrTransactionReviewDismissalInvalid = NewNormalError(NormalSubcategoryInsightsExplorer, 0, http.StatusBadRequest, "review dismissal is invalid")
)
