// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/stretchr/testify/assert"
)

func TestHTTPCORSMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		allowedOrigins  []string
		requestMethod   string
		requestHeaders  map[string]string
		expectedStatus  int
		expectHeaders   map[string]string
		expectNoHeaders []string
	}{
		{
			name:           "No Origin header",
			allowedOrigins: []string{"http://example.com"},
			requestMethod:  "GET",
			requestHeaders: map[string]string{},
			expectedStatus: http.StatusOK,
			expectNoHeaders: []string{
				"Access-Control-Allow-Origin",
			},
		},
		{
			name:           "Allowed Origin",
			allowedOrigins: []string{"http://example.com"},
			requestMethod:  "GET",
			requestHeaders: map[string]string{
				"Origin": "http://example.com",
			},
			expectedStatus: http.StatusOK,
			expectHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Credentials": "true",
				"Vary":                             "Origin",
			},
		},
		{
			name:           "Disallowed Origin",
			allowedOrigins: []string{"http://example.com"},
			requestMethod:  "GET",
			requestHeaders: map[string]string{
				"Origin": "http://evil.com",
			},
			expectedStatus: http.StatusOK, // Passed through, but no CORS headers
			expectNoHeaders: []string{
				"Access-Control-Allow-Origin",
			},
		},
		{
			name:           "Wildcard Origin (Secure Behavior)",
			allowedOrigins: []string{"*"},
			requestMethod:  "GET",
			requestHeaders: map[string]string{
				"Origin": "http://anywhere.com",
			},
			expectedStatus: http.StatusOK,
			expectHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
			expectNoHeaders: []string{
				"Access-Control-Allow-Credentials",
			},
		},
		{
			name:           "Wildcard and Specific Origin (Specific Match)",
			allowedOrigins: []string{"*", "http://example.com"},
			requestMethod:  "GET",
			requestHeaders: map[string]string{
				"Origin": "http://example.com",
			},
			expectedStatus: http.StatusOK,
			expectHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Credentials": "true",
			},
		},
		{
			name:           "OPTIONS Request (Preflight)",
			allowedOrigins: []string{"http://example.com"},
			requestMethod:  "OPTIONS",
			requestHeaders: map[string]string{
				"Origin": "http://example.com",
			},
			expectedStatus: http.StatusOK,
			expectHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "http://example.com",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS, PATCH",
				"Access-Control-Allow-Headers": "Content-Type, Authorization, X-API-Key, X-Requested-With, x-grpc-web, grpc-timeout, x-user-agent",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := middleware.NewHTTPCORSMiddleware(tt.allowedOrigins)

			// Mock handler
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := mw.Handler(next)

			req := httptest.NewRequest(tt.requestMethod, "/", nil)
			for k, v := range tt.requestHeaders {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			for k, v := range tt.expectHeaders {
				assert.Equal(t, v, resp.Header.Get(k), "Header %s mismatch", k)
			}

			for _, k := range tt.expectNoHeaders {
				assert.Empty(t, resp.Header.Get(k), "Header %s should be empty", k)
			}
		})
	}
}

func TestHTTPCORSMiddleware_Update(t *testing.T) {
	mw := middleware.NewHTTPCORSMiddleware([]string{"http://initial.com"})

	// Verify initial state
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://initial.com")
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	mw.Handler(next).ServeHTTP(w, req)
	assert.Equal(t, "http://initial.com", w.Result().Header.Get("Access-Control-Allow-Origin"))

	// Update configuration
	mw.Update([]string{"http://updated.com"})

	// Verify old origin is no longer allowed
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://initial.com")
	w = httptest.NewRecorder()
	mw.Handler(next).ServeHTTP(w, req)
	assert.Empty(t, w.Result().Header.Get("Access-Control-Allow-Origin"))

	// Verify new origin is allowed
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://updated.com")
	w = httptest.NewRecorder()
	mw.Handler(next).ServeHTTP(w, req)
	assert.Equal(t, "http://updated.com", w.Result().Header.Get("Access-Control-Allow-Origin"))
}
