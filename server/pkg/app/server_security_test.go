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

func TestCreateAuthMiddleware_Security(t *testing.T) {
	// Setup Application with no API Key
	app := &Application{
		SettingsManager: NewGlobalSettingsManager("", nil, nil),
		AuthManager:     auth.NewManager(),
	}

	// Create middleware
	// forcePrivateIPOnly = false, trustProxy = false
	mw := app.createAuthMiddleware(false, false)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(nextHandler)

	t.Run("Allow Localhost No API Key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Allow IPv6 Loopback No API Key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "[::1]:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Block LAN IP No API Key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.50:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "Access from non-localhost requires an API Key")
	})

	t.Run("Block Public IP No API Key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "8.8.8.8:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	// Setup Application WITH API Key
	appWithKey := &Application{
		SettingsManager: NewGlobalSettingsManager("secret-key", nil, nil),
		AuthManager:     auth.NewManager(),
	}
	mwWithKey := appWithKey.createAuthMiddleware(false, false)
	handlerWithKey := mwWithKey(nextHandler)

	t.Run("Allow LAN IP With API Key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.50:12345"
		req.Header.Set("X-API-Key", "secret-key")
		rec := httptest.NewRecorder()

		handlerWithKey.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Block LAN IP With Wrong API Key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.50:12345"
		req.Header.Set("X-API-Key", "wrong-key")
		rec := httptest.NewRecorder()

		handlerWithKey.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
