package api

import (
	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/errs"
	"github.com/James-HoneyBadger/honey-ledger/pkg/log"
	"github.com/James-HoneyBadger/honey-ledger/pkg/models"
	"github.com/James-HoneyBadger/honey-ledger/pkg/services"
)

// AIAssistantApi represents the AI assistant api
type AIAssistantApi struct {
	assistant *services.AIAssistantService
}

// Initialize an AI assistant api singleton instance
var (
	AIAssistant = &AIAssistantApi{
		assistant: services.AIAssistant,
	}
)

// AskHandler answers a natural language question about the user's finances.
func (a *AIAssistantApi) AskHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.AIAssistantAskRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Warnf(c, "[ai_assistant.AskHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	response, err := a.assistant.Ask(c, uid, &request)

	if err != nil {
		log.Errorf(c, "[ai_assistant.AskHandler] failed to answer question for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return response, nil
}
