package util_test

import (
	"testing"

	"github.com/mcpany/core/server/pkg/util"
)

func TestRedactDSN_Reproduction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Password with colon",
			input:    "postgres://user:pass:word@host/db",
			expected: "postgres://user:[REDACTED]@host/db",
		},
		{
			name:     "Password with at",
			input:    "postgres://user:p@ssword@host/db",
			expected: "postgres://user:[REDACTED]@host/db",
		},
		// "Password with hash" where hash IS the separator.
		// If user meant hash in password, they should encode it.
		// So we treat it as fragment separator.
		{
			name:     "Fragment separator in password",
			input:    "postgres://user:pass#word@host/db",
			expected: "postgres://user:[REDACTED]#word@host/db",
		},
		{
			name:     "Query separator in password",
			input:    "postgres://user:pass?word@host/db",
			expected: "postgres://user:[REDACTED]?word@host/db",
		},
        {
            name: "Redis invalid port",
            input: "redis://:password",
            expected: "redis://:[REDACTED]",
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := util.RedactDSN(tt.input)
			if got != tt.expected {
				t.Errorf("RedactDSN(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
