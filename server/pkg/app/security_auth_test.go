package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestLANAccessWithoutAPIKey(t *testing.T) {
	app := NewApplication()
	// No API Key set (default)
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	// Create middleware
	mw := app.createAuthMiddleware(false, false)

	// Dummy handler
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request from LAN IP
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.5:12345"

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Vulnerability: currently returns 200 OK because IsPrivateIP allows LAN
	// Expectation (Hardened): Should be 403 Forbidden
	assert.Equal(t, http.StatusForbidden, w.Code, "Expected 403 Forbidden for LAN access without API Key")
}

func TestQueryParamAPIKey(t *testing.T) {
	app := NewApplication()
	apiKey := "secret"
	app.SettingsManager = NewGlobalSettingsManager(apiKey, nil, nil)
	// Also set explicitAPIKey just in case
	app.explicitAPIKey = apiKey

	// Create middleware
	mw := app.createAuthMiddleware(false, false)

	// Dummy handler
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if we are authenticated
		_, ok := util.RemoteIPFromContext(r.Context())
		if ok {
			w.WriteHeader(http.StatusOK)
		} else {
			// Actually the middleware sets context if authenticated.
			// But createAuthMiddleware doesn't set RemoteIP in context, createAuthMiddleware calls next handler.
			// The middleware we are testing is AuthMiddleware.
			// Wait, the middleware logic:
			// ip := util.GetClientIP(r, trustProxy)
			// ctx := util.ContextWithRemoteIP(r.Context(), ip)
			// ...
			// So RemoteIP IS set.
			w.WriteHeader(http.StatusOK)
		}
	}))

	// Request with API Key in Query Param
	req := httptest.NewRequest(http.MethodGet, "/?api_key=secret", nil)
	req.RemoteAddr = "1.2.3.4:12345" // Public IP to force Auth check

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Vulnerability: currently returns 200 OK because Query Param is checked
	// Expectation (Hardened): Should be 401 Unauthorized (because we removed Query Param support)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected 401 Unauthorized for Query Param API Key")
}
