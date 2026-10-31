package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newTestWebContext(t *testing.T) (*WebContext, *httptest.ResponseRecorder) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	return &WebContext{Context: ginContext}, recorder
}

func TestGetTokenStringFromCookie_LegacyFallback(t *testing.T) {
	for _, legacyCookieParam := range legacyTokenCookieParams {
		webContext, _ := newTestWebContext(t)
		webContext.Request.AddCookie(&http.Cookie{Name: legacyCookieParam, Value: "legacy-token"})

		assert.Equal(t, "legacy-token", webContext.GetTokenStringFromCookie(), legacyCookieParam)
	}
}

func TestGetTokenStringFromCookie_PrefersCurrentCookie(t *testing.T) {
	webContext, _ := newTestWebContext(t)

	for _, legacyCookieParam := range legacyTokenCookieParams {
		webContext.Request.AddCookie(&http.Cookie{Name: legacyCookieParam, Value: "legacy-token"})
	}

	webContext.Request.AddCookie(&http.Cookie{Name: tokenCookieParam, Value: "current-token"})

	assert.Equal(t, "current-token", webContext.GetTokenStringFromCookie())
}

func TestSetTokenStringToCookie_ClearsLegacyCookie(t *testing.T) {
	webContext, recorder := newTestWebContext(t)
	webContext.SetTokenStringToCookie("", -1, "/", false)

	cookies := recorder.Result().Cookies()
	assert.Len(t, cookies, 1+len(legacyTokenCookieParams))

	values := make(map[string]string)
	for _, cookie := range cookies {
		values[cookie.Name] = cookie.Value
	}

	assert.Equal(t, "", values[tokenCookieParam])

	for _, legacyCookieParam := range legacyTokenCookieParams {
		assert.Equal(t, "", values[legacyCookieParam])
	}
}
