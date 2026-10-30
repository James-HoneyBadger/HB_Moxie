package provider

import (
	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/llm/data"
	"github.com/James-HoneyBadger/honey-ledger/pkg/settings"
)

// LargeLanguageModelProvider defines the structure of large language model provider
type LargeLanguageModelProvider interface {
	// GetJsonResponse returns the json response from the large language model provider
	GetJsonResponse(c core.Context, uid int64, currentLLMConfig *settings.LLMConfig, request *data.LargeLanguageModelRequest) (*data.LargeLanguageModelTextualResponse, error)
}
