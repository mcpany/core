// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"net/http"
	"testing"
)

func TestGetClientIP_XFF_Validation(t *testing.T) {
	tests := []struct {
		name       string
		xff        string
		expected   string
	}{
		{
			name:     "XFF with brackets",
			xff:      "[::1]",
			expected: "::1",
		},
		{
			name:     "XFF with zone index",
			xff:      "fe80::1%eth0",
			expected: "fe80::1",
		},
		{
			name:     "XFF with brackets and port",
			xff:      "[2001:db8::1]:8080",
			expected: "2001:db8::1",
		},
		{
			name:     "XFF with port only",
			xff:      "1.2.3.4:1234",
			expected: "1.2.3.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("X-Forwarded-For", tt.xff)
			// remoteAddr is irrelevant here as we trust proxy
			req.RemoteAddr = "127.0.0.1:12345"

			got := GetClientIP(req, true)
			if got != tt.expected {
				t.Errorf("GetClientIP() = %q, want %q", got, tt.expected)
			}

			// E2E Verification: Ensure the result is parseable by net.ParseIP
			// This confirms that downstream logic (like IsPrivateIP) will work correctly.
			ip := net.ParseIP(got)
			if ip == nil {
				t.Errorf("Resulting IP %q is not parseable by net.ParseIP", got)
			}
		})
	}
}
