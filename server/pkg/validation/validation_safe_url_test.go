// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSafeURL(t *testing.T) {
	testCases := []struct {
		name          string
		url           string
		allowedHosts  string // Comma-separated list for MCPANY_URL_ALLOW_LIST
		expectedError string
	}{
		// Safe URLs
		{"public http", "http://example.com", "", ""},
		{"public https", "https://google.com/search", "", ""},
		{"public ip", "http://8.8.8.8", "", ""},

		// Unsafe URLs
		{"localhost", "http://localhost:8080", "", "localhost is not allowed"},
		{"localhost case insensitive", "http://LOCALHOST:8080", "", "localhost is not allowed"},
		{"loopback ip", "http://127.0.0.1", "", "loopback IP \"127.0.0.1\" is not allowed"},
		{"ipv6 loopback", "http://[::1]", "", "loopback IP \"::1\" is not allowed"},
		{"private ip 10.x", "http://10.0.0.1", "", "private IP \"10.0.0.1\" is not allowed"},
		{"private ip 192.168.x", "http://192.168.1.1", "", "private IP \"192.168.1.1\" is not allowed"},
		{"private ip 172.16.x", "http://172.16.0.1", "", "private IP \"172.16.0.1\" is not allowed"},
		{"link local", "http://169.254.169.254", "", "link-local IP \"169.254.169.254\" is not allowed"},
		{"unspecified", "http://0.0.0.0", "", "unspecified IP \"0.0.0.0\" is not allowed"},

		// Allowed via Env Var
		{"allowed localhost", "http://localhost:8080", "localhost", ""},
		{"allowed private ip", "http://10.0.0.1", "10.0.0.1", ""},
		{"allowed multiple", "http://192.168.1.1", "localhost,192.168.1.1", ""},
        {"allowed mixed case", "http://example.local", "EXAMPLE.LOCAL", ""}, // Testing case insensitive allow list
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable
			if tc.allowedHosts != "" {
				os.Setenv("MCPANY_URL_ALLOW_LIST", tc.allowedHosts)
			} else {
				os.Unsetenv("MCPANY_URL_ALLOW_LIST")
			}
			defer os.Unsetenv("MCPANY_URL_ALLOW_LIST")

			err := IsSafeURL(tc.url)
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.expectedError)
			}
		})
	}
}
