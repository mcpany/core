// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/stretchr/testify/assert"
)

// TestAuthMiddleware_Security_Defaults validates that the auth middleware
// enforces strictly localhost access when no API key is configured, preventing
// accidental exposure to private networks (e.g. via Load Balancers).
func TestAuthMiddleware_Security_Defaults(t *testing.T) {
	// Setup
	app := &Application{
		SettingsManager: NewGlobalSettingsManager("", nil, nil),
		AuthManager:     auth.NewManager(),
	}

	// Helper to create a request with a specific remote address
	newReq := func(remoteAddr string) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
		req.RemoteAddr = remoteAddr
		return req
	}

	// Helper to create a request with auth header
	newReqWithAuth := func(remoteAddr, apiKey string) *http.Request {
		req := newReq(remoteAddr)
		req.Header.Set("X-API-Key", apiKey)
		return req
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Case 1: No API Key Configured
	// -----------------------------
	middleware := app.createAuthMiddleware(false, false)
	protected := middleware(handler)

	// 1a. Localhost (IPv4) -> Allowed
	w := httptest.NewRecorder()
	protected.ServeHTTP(w, newReq("127.0.0.1:1234"))
	assert.Equal(t, http.StatusOK, w.Code, "Localhost (127.0.0.1) should be allowed without API Key")

	// 1b. Localhost (IPv6) -> Allowed
	w = httptest.NewRecorder()
	protected.ServeHTTP(w, newReq("[::1]:1234"))
	assert.Equal(t, http.StatusOK, w.Code, "Localhost (::1) should be allowed without API Key")

	// 1c. Private Network (LAN) -> Blocked (CRITICAL CHECK)
	// This simulates a request from another machine on the LAN or a Load Balancer
	w = httptest.NewRecorder()
	protected.ServeHTTP(w, newReq("10.0.0.5:1234"))
	// Note: Before the fix, this might return 200 if the code uses IsPrivateIP.
	// We assert 403 because that's the DESIRED secure behavior.
	assert.Equal(t, http.StatusForbidden, w.Code, "Private IP (10.0.0.5) must be BLOCKED when no API Key is set")

	// 1d. Public Internet -> Blocked
	w = httptest.NewRecorder()
	protected.ServeHTTP(w, newReq("8.8.8.8:1234"))
	assert.Equal(t, http.StatusForbidden, w.Code, "Public IP should be blocked")

	// Case 2: API Key Configured
	// --------------------------
	apiKey := "s3cr3t-k3y"
	app.SettingsManager.Update(nil, apiKey)
	middleware = app.createAuthMiddleware(false, false)
	protected = middleware(handler)

	// 2a. Any IP (Private) without Key -> Unauthorized
	w = httptest.NewRecorder()
	protected.ServeHTTP(w, newReq("10.0.0.5:1234"))
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Request without key should be Unauthorized when key is configured")

	// 2b. Any IP (Localhost) without Key -> Unauthorized
	// Even localhost requires key if one is set! (Standard security practice)
	w = httptest.NewRecorder()
	protected.ServeHTTP(w, newReq("127.0.0.1:1234"))
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Localhost request without key should be Unauthorized when key is configured")

	// 2c. Any IP with Correct Key -> Allowed
	w = httptest.NewRecorder()
	protected.ServeHTTP(w, newReqWithAuth("10.0.0.5:1234", apiKey))
	assert.Equal(t, http.StatusOK, w.Code, "Request with correct key should be Allowed")
}
