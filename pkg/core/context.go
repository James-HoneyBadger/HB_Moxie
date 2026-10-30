package core

import "context"

// Context is the base context of Honey Ledger
type Context interface {
	context.Context
	GetContextId() string
	GetClientLocale() string
}
