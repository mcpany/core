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
		allowed        []string
		remoteAddr     string
		expectedStatus int
	}{
		{
			name:           "Empty allowlist allows all",
			allowed:        []string{},
			remoteAddr:     "192.168.1.1:1234",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allow specific IP",
			allowed:        []string{"192.168.1.1"},
			remoteAddr:     "192.168.1.1:1234",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Deny different IP",
			allowed:        []string{"192.168.1.1"},
			remoteAddr:     "192.168.1.2:1234",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Allow CIDR range",
			allowed:        []string{"192.168.1.0/24"},
			remoteAddr:     "192.168.1.50:5678",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Deny outside CIDR range",
			allowed:        []string{"192.168.1.0/24"},
			remoteAddr:     "192.168.2.1:1234",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Allow IPv6",
			allowed:        []string{"::1"},
			remoteAddr:     "[::1]:1234",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allow IPv6 CIDR",
			allowed:        []string{"2001:db8::/32"},
			remoteAddr:     "[2001:db8::1]:1234",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Deny IPv6 outside CIDR",
			allowed:        []string{"2001:db8::/32"},
			remoteAddr:     "[2001:db9::1]:1234",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid RemoteAddr format (no port)",
			allowed:        []string{"127.0.0.1"},
			remoteAddr:     "127.0.0.1",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewIPAllowlistMiddleware(tt.allowed)
			require.NoError(t, err)

			handler := m.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestNewIPAllowlistMiddleware_InvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		allowed []string
		wantErr bool
	}{
		{
			name:    "Invalid IP",
			allowed: []string{"invalid-ip"},
			wantErr: true,
		},
		{
			name:    "Invalid CIDR",
			allowed: []string{"192.168.1.1/500"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewIPAllowlistMiddleware(tt.allowed)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
