package middlewares

import (
	"github.com/James-HoneyBadger/ratel-ledger/pkg/core"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/errs"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/settings"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/utils"
)

// APITokenIpLimit limits API token access based on IP address
func APITokenIpLimit(config *settings.Config) core.MiddlewareHandlerFunc {
	return func(c *core.WebContext) {
		claims := c.GetTokenClaims()

		if claims == nil {
			c.Next()
			return
		}

		if claims.Type != core.USER_TOKEN_TYPE_API {
			c.Next()
			return
		}

		if len(config.APITokenAllowedRemoteIPs) < 1 {
			c.Next()
			return
		}

		for i := 0; i < len(config.APITokenAllowedRemoteIPs); i++ {
			if config.APITokenAllowedRemoteIPs[i].Match(c.ClientIP()) {
				c.Next()
				return
			}
		}

		utils.PrintJsonErrorResult(c, errs.ErrIPForbidden)
	}
}
