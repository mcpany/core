// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPAllowlistMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		allowedIPs     []string
		remoteAddr     string
		expectedStatus int
	}{
		{
			name:           "Empty allowlist allows all",
			allowedIPs:     []string{},
			remoteAddr:     "192.168.1.1:1234",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allow specific IP",
			allowedIPs:     []string{"192.168.1.1"},
			remoteAddr:     "192.168.1.1:5678",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Deny unlisted IP",
			allowedIPs:     []string{"192.168.1.1"},
			remoteAddr:     "192.168.1.2:1234",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Allow CIDR range",
			allowedIPs:     []string{"192.168.1.0/24"},
			remoteAddr:     "192.168.1.50:9000",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Deny outside CIDR range",
			allowedIPs:     []string{"192.168.1.0/24"},
			remoteAddr:     "192.168.2.1:1234",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "IPv6 allow",
			allowedIPs:     []string{"::1"},
			remoteAddr:     "[::1]:1234",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IPv6 deny",
			allowedIPs:     []string{"::1"},
			remoteAddr:     "[::2]:1234",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewIPAllowlistMiddleware(tt.allowedIPs)
			require.NoError(t, err)

			handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestIPAllowlistMiddleware_InvalidConfig(t *testing.T) {
	_, err := NewIPAllowlistMiddleware([]string{"invalid-ip"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid IP or CIDR")
}
