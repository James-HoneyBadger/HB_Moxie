package middlewares

import (
	"github.com/James-HoneyBadger/honey-ledger/pkg/core"
	"github.com/James-HoneyBadger/honey-ledger/pkg/settings"
)

// AmapApiProxyAuthCookie adds amap api proxy auth cookie to cookies in response
func AmapApiProxyAuthCookie(c *core.WebContext, config *settings.Config) {
	token := c.GetTextualToken()
	c.SetTokenStringToCookie(token, int(config.TokenExpiredTime), "/_AMapService", false)
}
