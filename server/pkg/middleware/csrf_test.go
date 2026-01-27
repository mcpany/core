// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCSRFMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		allowedOrigins []string
		headers        map[string]string
		host           string
		wantStatus     int
	}{
		{
			name:       "Safe Method GET",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Safe Method OPTIONS",
			method:     http.MethodOptions,
			wantStatus: http.StatusOK,
		},
		{
			name:   "POST with Authorization",
			method: http.MethodPost,
			headers: map[string]string{
				"Authorization": "Bearer token",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "POST with X-API-Key",
			method: http.MethodPost,
			headers: map[string]string{
				"X-API-Key": "key",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:           "POST with Valid Origin",
			method:         http.MethodPost,
			allowedOrigins: []string{"http://example.com"},
			headers: map[string]string{
				"Origin": "http://example.com",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:           "POST with Invalid Origin",
			method:         http.MethodPost,
			allowedOrigins: []string{"http://example.com"},
			headers: map[string]string{
				"Origin": "http://evil.com",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:           "POST with Same Origin (Matches Host)",
			method:         http.MethodPost,
			allowedOrigins: []string{}, // Empty allowed
			headers: map[string]string{
				"Origin": "http://localhost:8080",
			},
			host:       "localhost:8080",
			wantStatus: http.StatusOK,
		},
		{
			name:           "POST with Same Origin (Matches Host) 2",
			method:         http.MethodPost,
			allowedOrigins: []string{},
			headers: map[string]string{
				"Origin": "https://myapp.com",
			},
			host:       "myapp.com",
			wantStatus: http.StatusOK,
		},
		{
			name:           "POST with Mismatched Host and Origin",
			method:         http.MethodPost,
			allowedOrigins: []string{},
			headers: map[string]string{
				"Origin": "http://evil.com",
			},
			host:       "localhost:8080",
			wantStatus: http.StatusForbidden,
		},
		{
			name:           "POST with Valid Referer",
			method:         http.MethodPost,
			allowedOrigins: []string{"http://example.com"},
			headers: map[string]string{
				"Referer": "http://example.com/page",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:           "POST with Invalid Referer",
			method:         http.MethodPost,
			allowedOrigins: []string{"http://example.com"},
			headers: map[string]string{
				"Referer": "http://evil.com/page",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:           "POST with Same Origin Referer",
			method:         http.MethodPost,
			allowedOrigins: []string{},
			headers: map[string]string{
				"Referer": "http://localhost:8080/ui",
			},
			host:       "localhost:8080",
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST with No Origin/Referer (Non-Browser)",
			method:     http.MethodPost,
			wantStatus: http.StatusOK,
		},
		{
			name:           "POST with Wildcard Allowed Origin",
			method:         http.MethodPost,
			allowedOrigins: []string{"*"},
			headers: map[string]string{
				"Origin": "http://evil.com",
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCSRFMiddleware(tt.allowedOrigins)
			handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, "http://localhost:8080/api", nil)
			if tt.host != "" {
				req.Host = tt.host
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
