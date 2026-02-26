// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"strings"
	"testing"
)

func TestValidateSafePathAndInjection_LoopbackShorthand(t *testing.T) {
	// Ensure strict environment for testing
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "")

	tests := []struct {
		val         string
		shouldError bool
		errorMsg    string
	}{
		{"127.0.0.1", true, "unsafe IP argument"}, // Blocked by IsSafeIP
		{"127.1", true, "loopback shorthand address is not allowed"}, // Blocked by IsLoopbackShorthand
		{"127.0.1", true, "loopback shorthand address is not allowed"},
		{"127.255.255.255", true, "unsafe IP argument"}, // Blocked by IsSafeIP
		{"127.2.3.4", true, "unsafe IP argument"},
		{"127.123", true, "loopback shorthand address is not allowed"},
		{"127.", true, "loopback shorthand address is not allowed"}, // Technically shorthand
		{"10.1", false, ""}, // Private IP shorthand, but IsSafeIP misses it (nil) and IsLoopbackShorthand misses it. Allowed for now to avoid FP.
		{"0", false, ""},    // Allowed for now
		{"127.txt", false, ""}, // Contains letters, not shorthand
		{"127.0.0.1.txt", false, ""}, // Contains letters
		{"not-an-ip", false, ""},
		{"localhost", true, "localhost is not allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			err := validateSafePathAndInjection(tt.val, false, "curl")
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for %q, got nil", tt.val)
				} else if !strings.Contains(err.Error(), tt.errorMsg) && !strings.Contains(err.Error(), "loopback address is not allowed") {
					// IsSafeIP returns "loopback address is not allowed"
					// Our new check returns "loopback shorthand address is not allowed"
					t.Errorf("expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error for %q, got %v", tt.val, err)
				}
			}
		})
	}
}
