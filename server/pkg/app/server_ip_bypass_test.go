// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware_IPBypass(t *testing.T) {
	// Setup application with no API key (triggers private IP enforcement)
	app := NewApplication()
	// Initialize SettingsManager to avoid nil pointer
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	// Create middleware
	// forcePrivateIPOnly = false (default behavior when API key is checked but missing)
	// But in createAuthMiddleware:
	// if !forcePrivateIPOnly && apiKey != "" { ... } else { ... check IsPrivateIP ... }
	// So passing false + empty API key triggers the Sentinel Security check.
	middleware := app.createAuthMiddleware(false)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name       string
		remoteAddr string
		wantStatus int
	}{
		{
			name:       "IPv4 Loopback",
			remoteAddr: "127.0.0.1:12345",
			wantStatus: http.StatusOK,
		},
		{
			name:       "IPv4 Private",
			remoteAddr: "192.168.1.1:12345",
			wantStatus: http.StatusOK,
		},
		{
			name:       "IPv6 Loopback",
			remoteAddr: "[::1]:12345",
			wantStatus: http.StatusOK,
		},
		{
			name:       "IPv4 Public",
			remoteAddr: "8.8.8.8:12345",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "IPv4-Compatible Loopback (Bypass Attempt)",
			remoteAddr: "[::127.0.0.1]:12345",
			wantStatus: http.StatusOK, // Should be allowed because it IS loopback (previously failed)
		},
		{
			name:       "IPv4-Compatible Private (Bypass Attempt)",
			remoteAddr: "[::192.168.1.1]:12345",
			wantStatus: http.StatusOK, // Should be allowed because it IS private (previously failed)
		},
		{
			name:       "IPv4-Compatible Public (Attack)",
			remoteAddr: "[::8.8.8.8]:12345",
			wantStatus: http.StatusForbidden, // Should be blocked
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v, want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}
