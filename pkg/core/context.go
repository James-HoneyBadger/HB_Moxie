package core

import "context"

// Context is the base context of Ratel Ledger
type Context interface {
	context.Context
	GetContextId() string
	GetClientLocale() string
}
