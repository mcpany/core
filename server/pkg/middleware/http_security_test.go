// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/stretchr/testify/assert"
)

func TestHTTPSecurityHeadersMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.HTTPSecurityHeadersMiddleware(nextHandler)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	assert.Equal(t, "SAMEORIGIN", resp.Header.Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", resp.Header.Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", resp.Header.Get("Referrer-Policy"))
	assert.Equal(t, "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; script-src 'self'; connect-src 'self'", resp.Header.Get("Content-Security-Policy"))
	assert.Equal(t, "max-age=63072000; includeSubDomains", resp.Header.Get("Strict-Transport-Security"))
	assert.Equal(t, "geolocation=(), camera=(), microphone=(), payment=(), usb=(), vr=()", resp.Header.Get("Permissions-Policy"))
	assert.Equal(t, "no-store, no-cache, must-revalidate, proxy-revalidate", resp.Header.Get("Cache-Control"))
	assert.Equal(t, "no-cache", resp.Header.Get("Pragma"))
	assert.Equal(t, "0", resp.Header.Get("Expires"))
	assert.Equal(t, "", resp.Header.Get("Server"))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
