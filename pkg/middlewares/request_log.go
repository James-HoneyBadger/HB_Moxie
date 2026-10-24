package middlewares

import (
	"net/url"
	"strings"
	"time"

	"github.com/James-HoneyBadger/ratel-ledger/pkg/core"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/log"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/utils"
)

const redactedQueryValue = "***"

var sensitiveQueryParameterNames = map[string]bool{
	"token":        true,
	"access_token": true,
	"id_token":     true,
	"password":     true,
	"code":         true,
	"state":        true,
}

// RequestLog logs the http request log
func RequestLog(c *core.WebContext) {
	start := time.Now()
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery

	c.Next()

	now := time.Now()

	statusCode := c.Writer.Status()
	errorCode := int32(0)

	userId := "-"
	claims := c.GetTokenClaims()
	err := c.GetResponseError()

	clientIP := c.ClientIP()
	method := c.Request.Method

	if claims != nil {
		userId = utils.Int64ToString(claims.Uid)
	}

	if err != nil {
		errorCode = err.Code()
	}

	if query != "" {
		path = path + "?" + redactSensitiveQuery(query)
	}

	cost := now.Sub(start).Nanoseconds() / 1e6

	log.Requestf(c, "%d %d %s %s %s %s %dms", statusCode, errorCode, userId, clientIP, method, path, cost)
}

// redactSensitiveQuery removes credentials that are passed in the query string
// (for example image and proxy URLs) from request logs.
func redactSensitiveQuery(rawQuery string) string {
	values, err := url.ParseQuery(rawQuery)

	if err != nil {
		return redactedQueryValue
	}

	for key, items := range values {
		if !sensitiveQueryParameterNames[strings.ToLower(key)] {
			continue
		}

		for i := range items {
			items[i] = redactedQueryValue
		}

		values[key] = items
	}

	return values.Encode()
}
