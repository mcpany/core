// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSafeURL_SSRF(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		// Allowed public IPs
		{"Public IP", "http://8.8.8.8", false},
		{"Public Domain", "http://example.com", false}, // Assuming example.com resolves to public IP

		// Private IPs (Blocked)
		{"Localhost", "http://localhost", true},
		{"127.0.0.1", "http://127.0.0.1", true},
		{"IPv6 Loopback", "http://[::1]", true},
		{"Private 10.x", "http://10.0.0.1", true},
		{"Private 172.16.x", "http://172.16.0.1", true},
		{"Private 192.168.x", "http://192.168.1.1", true},
		{"Metadata Service", "http://169.254.169.254", true},
		{"0.0.0.0", "http://0.0.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsSafeURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for URL: %s", tt.url)
			} else {
				assert.NoError(t, err, "Expected no error for URL: %s", tt.url)
			}
		})
	}
}
