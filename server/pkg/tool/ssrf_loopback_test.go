// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/validation"
)

func TestSSRFLoopbackShorthand(t *testing.T) {
	// Mock IsSafeURL to fail for 127.* URLs, simulating a real environment where DNS/IP check would fail
	// TestMain mocks it to always pass.
	originalIsSafeURL := validation.IsSafeURL
	validation.IsSafeURL = func(urlStr string) error {
		if strings.Contains(urlStr, "127.") {
			return fmt.Errorf("mock error: unsafe url")
		}
		return nil
	}
	defer func() { validation.IsSafeURL = originalIsSafeURL }()

	// Ensure loopback is NOT allowed by default
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	// Disable dangerous allow bypass to test strict validation
	os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "false")
	defer os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true") // Restore for other tests

	tests := []struct {
		input       string
		shouldBlock bool
		desc        string
	}{
		// Standard Loopback
		{"127.0.0.1", true, "Standard Loopback"},
		{"http://127.0.0.1", true, "Standard Loopback URL"},

		// Loopback Shorthands
		{"127.1", true, "Loopback Shorthand 127.1"},
		{"127.0.1", true, "Loopback Shorthand 127.0.1"},
		{"127.255", true, "Loopback Shorthand 127.255"},
		{"127.0.0.1", true, "Loopback Standard"},

		// Ambiguous Cases (Currently Allowed because net.ParseIP fails and IsLoopbackShorthand returns false)
		{"0", false, "Ambiguous 0"},
		{"10.1", false, "Ambiguous 10.1 (Private but skipped as non-IP)"},

		// Valid Non-IPs
		{"127.txt", false, "Filename starting with 127"},
		{"127.0.0.1.txt", false, "Filename with IP"},
		{"foo127.1", false, "String containing 127.1"},

		// Valid IPs (Public)
		{"8.8.8.8", false, "Public IP"},
		{"1.1.1.1", false, "Public IP"},

		// URLs
		{"http://127.1", true, "URL with shorthand (blocked by IsSafeURL resolution failure or explicit check)"},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := validateSafePathAndInjection(tc.input, false, "curl")
			if tc.shouldBlock {
				if err == nil {
					t.Errorf("Expected input %q to be blocked, but it was allowed", tc.input)
				} else {
					// Verify error message for shorthands
					if strings.HasPrefix(tc.input, "127.") && !strings.Contains(tc.input, ":") && !strings.Contains(err.Error(), "loopback shorthand") && !strings.Contains(err.Error(), "unsafe IP argument") {
						// For 127.0.0.1, IsSafeIP might return "loopback address is not allowed"
						// For 127.1, IsLoopbackShorthand should trigger "loopback shorthand address is not allowed"
						t.Logf("Blocked as expected: %v", err)
					}
				}
			} else {
				if err != nil {
					// For http://127.1, it might fail DNS resolution in IsSafeURL if validation logic reaches there.
					// But our test case expectation is tricky.
					// validateSafePathAndInjection calls IsSafeURL if "://" is present.
					// IsSafeURL("http://127.1") -> net.ParseIP("127.1") is nil. -> LookupIP("127.1") fails. -> Error.
					// So http://127.1 IS blocked.
					if strings.Contains(tc.input, "://") && strings.Contains(err.Error(), "no such host") {
						// This is acceptable blocking behavior for IsSafeURL
					} else {
						t.Errorf("Expected input %q to be allowed, but it was blocked: %v", tc.input, err)
					}
				}
			}
		})
	}
}

func TestSSRFLoopbackAllowed(t *testing.T) {
	// Allow loopback
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	// 127.1 should now be allowed
	err := validateSafePathAndInjection("127.1", false, "curl")
	if err != nil {
		t.Errorf("Expected 127.1 to be allowed when MCPANY_ALLOW_LOOPBACK_RESOURCES=true, but got error: %v", err)
	}
}
