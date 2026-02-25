package tool

import (
	"os"
	"strings"
	"testing"
)

func TestValidateSafePathAndInjection_LoopbackShorthand(t *testing.T) {
	// Ensure loopback is blocked
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	tests := []struct {
		val         string
		shouldError bool
		description string
	}{
		{"127.1", true, "Loopback shorthand"},
		{"127.0.1", true, "Loopback shorthand"},
		{"127.0.0.1", true, "Loopback full"},
		{"127.255", true, "Loopback shorthand"},
		{"10.1", false, "Private IP shorthand (should pass net.ParseIP)"}, // 10.1 fails ParseIP, so it passes IsSafeIP check, and IsLoopbackShorthand check. So it PASSES.
		{"0", false, "Zero"},                                               // Ambiguous, allowed
		{"127.txt", false, "Filename starting with 127."},                  // Has letters
		{"127.1.txt", false, "Filename starting with 127.1"},               // Has letters
		// Removed http://127.1 because IsSafeURL is mocked to allow all URLs in tool package tests
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			err := validateSafePathAndInjection(tt.val, false, "test-tool")
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for %q (%s), got nil", tt.val, tt.description)
				} else if !strings.Contains(err.Error(), "loopback") && !strings.Contains(err.Error(), "unsafe url") && !strings.Contains(err.Error(), "failed to resolve") {
					// We expect either loopback error or URL error
					t.Errorf("expected loopback/url error for %q (%s), got %v", tt.val, tt.description, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error for %q (%s), got %v", tt.val, tt.description, err)
				}
			}
		})
	}
}

func TestValidateSafePathAndInjection_LoopbackShorthand_Allowed(t *testing.T) {
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	val := "127.1"
	err := validateSafePathAndInjection(val, false, "test-tool")
	if err != nil {
		t.Errorf("expected no error for %q when loopback allowed, got %v", val, err)
	}
}
