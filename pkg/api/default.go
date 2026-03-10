package api

import (
	"github.com/James-HoneyBadger/HB_Moxie/pkg/core"
	"github.com/James-HoneyBadger/HB_Moxie/pkg/errs"
)

// DefaultApi represents default api
type DefaultApi struct{}

// Initialize a default api singleton instance
var (
	Default = &DefaultApi{}
)

// ApiNotFound returns api not found error
func (a *DefaultApi) ApiNotFound(c *core.WebContext) (any, *errs.Error) {
	return nil, errs.ErrApiNotFound
}

// MethodNotAllowed returns method not allowed error
func (a *DefaultApi) MethodNotAllowed(c *core.WebContext) (any, *errs.Error) {
	return nil, errs.ErrMethodNotAllowed
}
