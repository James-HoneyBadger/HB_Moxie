package errs

import "net/http"

// Error codes related to the AI assistant
var (
	ErrAIAssistantNotEnabled      = NewNormalError(NormalSubcategoryAIAssistant, 0, http.StatusBadRequest, "ai assistant is not enabled")
	ErrAIAssistantNotConfigured   = NewNormalError(NormalSubcategoryAIAssistant, 1, http.StatusBadRequest, "ai assistant model is not configured")
	ErrAIAssistantQuestionInvalid = NewNormalError(NormalSubcategoryAIAssistant, 2, http.StatusBadRequest, "ai assistant question is invalid")
	ErrAIAssistantEmptyAnswer     = NewNormalError(NormalSubcategoryAIAssistant, 3, http.StatusBadGateway, "the ai assistant returned an empty answer")
)
