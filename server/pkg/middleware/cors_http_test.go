// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPCORSMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		allowedOrigins     []string
		requestOrigin      string
		expectedOrigin     string
		expectedCreds      string
		expectAllowHeaders bool
	}{
		{
			name:           "No Origin Header",
			allowedOrigins: []string{"*"},
			requestOrigin:  "",
			expectedOrigin: "",
		},
		{
			name:           "Allowed Origin *",
			allowedOrigins: []string{"*"},
			requestOrigin:  "http://example.com",
			expectedOrigin: "*",
			expectedCreds:  "", // Should NOT be present
		},
		{
			name:           "Specific Allowed Origin",
			allowedOrigins: []string{"http://example.com"},
			requestOrigin:  "http://example.com",
			expectedOrigin: "http://example.com",
			expectedCreds:  "true",
		},
		{
			name:           "Disallowed Origin",
			allowedOrigins: []string{"http://example.com"},
			requestOrigin:  "http://evil.com",
			expectedOrigin: "",
		},
		{
			name:           "Multiple Allowed Origins - Match",
			allowedOrigins: []string{"http://foo.com", "http://bar.com"},
			requestOrigin:  "http://bar.com",
			expectedOrigin: "http://bar.com",
			expectedCreds:  "true",
		},
		{
			name:           "Multiple Allowed Origins - No Match",
			allowedOrigins: []string{"http://foo.com", "http://bar.com"},
			requestOrigin:  "http://baz.com",
			expectedOrigin: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewHTTPCORSMiddleware(tt.allowedOrigins)
			handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedOrigin, rec.Header().Get("Access-Control-Allow-Origin"))
			if tt.expectedOrigin != "" {
				assert.Equal(t, tt.expectedCreds, rec.Header().Get("Access-Control-Allow-Credentials"))
				if tt.expectedCreds == "true" {
					assert.Equal(t, "Origin", rec.Header().Get("Vary"))
				}
				assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Methods"))
				assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Headers"))
			} else {
				assert.Empty(t, rec.Header().Get("Access-Control-Allow-Credentials"))
			}
		})
	}
}

func TestHTTPCORSMiddleware_Preflight(t *testing.T) {
	m := NewHTTPCORSMiddleware([]string{"http://example.com"})
	handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "http://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "POST")
}
