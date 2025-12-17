// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIPAllowlistMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		allowedIPs     []string
		remoteAddr     string
		expectedStatus int
	}{
		{
			name:           "Empty allowlist - Allow All",
			allowedIPs:     []string{},
			remoteAddr:     "1.2.3.4:1234",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed IP",
			allowedIPs:     []string{"127.0.0.1"},
			remoteAddr:     "127.0.0.1:5000",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed CIDR",
			allowedIPs:     []string{"10.0.0.0/8"},
			remoteAddr:     "10.1.2.3:1234",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Denied IP",
			allowedIPs:     []string{"127.0.0.1"},
			remoteAddr:     "1.2.3.4:1234",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Denied CIDR",
			allowedIPs:     []string{"192.168.1.0/24"},
			remoteAddr:     "192.168.2.1:1234",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid IP in config - Ignore and Deny",
			allowedIPs:     []string{"invalid-ip"},
			remoteAddr:     "127.0.0.1:1234",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "IPv6 Allowed",
			allowedIPs:     []string{"::1"},
			remoteAddr:     "[::1]:1234",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IPv6 Denied",
			allowedIPs:     []string{"::1"},
			remoteAddr:     "[::2]:1234",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := IPAllowlistMiddleware(tt.allowedIPs)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
