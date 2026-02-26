package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSafePathAndInjection_LoopbackShorthand(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		wantErr bool
	}{
		{"Valid File", "file.txt", false},
		{"Loopback Full", "127.0.0.1", true}, // Handled by IsSafeIP
		{"Loopback Shorthand 1", "127.1", true},
		{"Loopback Shorthand 2", "127.0.1", true},
		{"Loopback Shorthand 3", "127.255", true},
		{"Loopback Shorthand 4", "127.123.456", true}, // Still starts with 127.
		{"Ambiguous 0", "0", false}, // We decided to allow 0 for now
		{"Ambiguous 10.1", "10.1", false},
		{"Not Loopback", "128.1", false}, // Valid IP or not, but not loopback shorthand
		{"Valid Hostname", "127.example.com", false},
		{"Valid Filename with 127", "127.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSafePathAndInjection(tt.val, false, "test")
			if tt.wantErr {
				assert.Error(t, err, "expected error for %s", tt.val)
			} else {
				assert.NoError(t, err, "expected no error for %s", tt.val)
			}
		})
	}
}
