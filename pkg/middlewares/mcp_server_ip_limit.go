package middlewares

import (
	"github.com/James-HoneyBadger/HB_Moxie/pkg/core"
	"github.com/James-HoneyBadger/HB_Moxie/pkg/errs"
	"github.com/James-HoneyBadger/HB_Moxie/pkg/settings"
	"github.com/James-HoneyBadger/HB_Moxie/pkg/utils"
)

// MCPServerIpLimit limits access to the MCP server based on IP address.
func MCPServerIpLimit(config *settings.Config) core.MiddlewareHandlerFunc {
	return func(c *core.WebContext) {
		if len(config.MCPAllowedRemoteIPs) < 1 {
			c.Next()
			return
		}

		for i := 0; i < len(config.MCPAllowedRemoteIPs); i++ {
			if config.MCPAllowedRemoteIPs[i].Match(c.ClientIP()) {
				c.Next()
				return
			}
		}

		utils.PrintJsonErrorResult(c, errs.ErrIPForbidden)
	}
}
