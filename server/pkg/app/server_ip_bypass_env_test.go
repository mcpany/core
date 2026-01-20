// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/util"
)

func TestAuthMiddleware_IPBypass_EnvOverride(t *testing.T) {
	// Save current env and restore after test
	originalVal := os.Getenv("MCPANY_ALLOW_LAN_NO_AUTH")
	defer os.Setenv("MCPANY_ALLOW_LAN_NO_AUTH", originalVal)

	// Set env to allow LAN
	os.Setenv("MCPANY_ALLOW_LAN_NO_AUTH", util.TrueStr)

	// Setup application with no API key
	app := NewApplication()
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	// Create middleware
	middleware := app.createAuthMiddleware(false, false)

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
			name:       "IPv4 Private (Allowed by Env)",
			remoteAddr: "192.168.1.1:12345",
			wantStatus: http.StatusOK, // Should be allowed now
		},
		{
			name:       "IPv4 Public (Still Blocked)",
			remoteAddr: "8.8.8.8:12345",
			wantStatus: http.StatusForbidden,
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
