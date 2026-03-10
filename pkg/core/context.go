package core

import "context"

// Context is the base context of HB Moxie
type Context interface {
	context.Context
	GetContextId() string
	GetClientLocale() string
}
