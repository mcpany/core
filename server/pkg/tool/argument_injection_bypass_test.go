package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSafePathAndInjection_Bypass(t *testing.T) {
	// These inputs should be BLOCKED.
	tests := []struct {
		name     string
		input    string
		isDocker bool
	}{
		{
			name:     "Argument Injection with Leading Space",
			input:    " -dangerous",
			isDocker: false,
		},
		{
			name:     "Path Traversal with Leading Space",
			input:    " ../etc/passwd",
			isDocker: false,
		},
		{
			name:     "Local File Access with Leading Space",
			input:    " /etc/passwd", // Absolute path check
			isDocker: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSafePathAndInjection(tt.input, tt.isDocker)
			assert.Error(t, err, "Should block input: %q", tt.input)
		})
	}
}
