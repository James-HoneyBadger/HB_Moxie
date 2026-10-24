package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactSensitiveQuery_RedactsToken(t *testing.T) {
	actual := redactSensitiveQuery("token=secret-token&size=large")

	assert.Contains(t, actual, "token=%2A%2A%2A")
	assert.NotContains(t, actual, "secret-token")
	assert.Contains(t, actual, "size=large")
}

func TestRedactSensitiveQuery_RedactsMultipleValues(t *testing.T) {
	actual := redactSensitiveQuery("token=first&token=second&code=oauth-code&state=oauth-state")

	assert.NotContains(t, actual, "first")
	assert.NotContains(t, actual, "second")
	assert.NotContains(t, actual, "oauth-code")
	assert.NotContains(t, actual, "oauth-state")
}

func TestRedactSensitiveQuery_KeepsOrdinaryParameters(t *testing.T) {
	actual := redactSensitiveQuery("zoomLevel=12&coordinateX=2048&fileName=1.png")

	assert.Equal(t, "coordinateX=2048&fileName=1.png&zoomLevel=12", actual)
}

func TestRedactSensitiveQuery_InvalidQueryIsNotLogged(t *testing.T) {
	actual := redactSensitiveQuery("token=%zz")

	assert.Equal(t, redactedQueryValue, actual)
}
