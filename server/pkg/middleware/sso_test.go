// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSOMiddleware(t *testing.T) {
	config := SSOConfig{
		Enabled: true,
		IDPURL:  "https://idp.example.com",
	}

	handler := SSOMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test No Auth
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test ID Header
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("X-MCP-Identity", "alice")
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test Bearer Token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-mock-token")
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
