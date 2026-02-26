// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSafePathAndInjection_LoopbackShorthand(t *testing.T) {
	// Ensure loopback is blocked by default
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	tests := []struct {
		name      string
		val       string
		shouldErr bool
		errContains string
	}{
		{
			name:      "Loopback shorthand 127.1",
			val:       "127.1",
			shouldErr: true,
			errContains: "loopback shorthand address is not allowed",
		},
		{
			name:      "Loopback shorthand 127.0.1",
			val:       "127.0.1",
			shouldErr: true,
			errContains: "loopback shorthand address is not allowed",
		},
		{
			name:      "Loopback shorthand 127.255",
			val:       "127.255",
			shouldErr: true,
			errContains: "loopback shorthand address is not allowed",
		},
		{
			name:      "Standard Loopback 127.0.0.1",
			val:       "127.0.0.1",
			shouldErr: true,
			// Caught by shorthand check first now, or IsSafeIP. Both are fine.
			errContains: "not allowed",
		},
        {
			name:      "Loopback shorthand with port 127.1:80",
			val:       "127.1:80",
			shouldErr: true,
			errContains: "loopback shorthand address is not allowed",
		},
        {
			name:      "Standard Loopback with port 127.0.0.1:8080",
			val:       "127.0.0.1:8080",
			shouldErr: true,
			errContains: "loopback shorthand address is not allowed", // net.ParseIP fails on port, shorthand check catches it
		},
		{
			name:      "Ambiguous 0 (Allowed)",
			val:       "0",
			shouldErr: false, // Net.ParseIP fails, not 127., so allowed
		},
		{
			name:      "Ambiguous 10.1 (Allowed)",
			val:       "10.1",
			shouldErr: false, // Not 127., allowed (could be version)
		},
		{
			name:      "Filename 127.txt (Allowed)",
			val:       "127.txt",
			shouldErr: false, // Contains letters
		},
		{
			name:      "Public IP 1.1.1.1 (Allowed)",
			val:       "1.1.1.1",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSafePathAndInjection(tt.val, false, "curl")
			if tt.shouldErr {
				if assert.Error(t, err) {
                    if tt.errContains != "" {
					    assert.Contains(t, err.Error(), tt.errContains)
                    }
                }
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSafePathAndInjection_LoopbackAllowed(t *testing.T) {
	// Allow loopback
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	val := "127.1"
	err := validateSafePathAndInjection(val, false, "curl")
	assert.NoError(t, err, "Should allow 127.1 when loopback is allowed")
}
