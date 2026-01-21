package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactDSN_Fix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "fallback: password with colon (no scheme)",
			input:    "user:pass:word@tcp(host:3306)/db",
			expected: "user:[REDACTED]@tcp(host:3306)/db",
		},
		{
			name:     "fallback: password with at (no scheme)",
			input:    "user:pass@word@tcp(host:3306)/db",
			expected: "user:[REDACTED]@tcp(host:3306)/db",
		},
		{
			name:     "fallback: password with colon and at (no scheme)",
			input:    "user:pass:word@place@tcp(host:3306)/db",
			expected: "user:[REDACTED]@tcp(host:3306)/db",
		},
		{
			name:     "fallback: invalid url encoding triggers fallback",
			input:    "postgres://user:p%ssword@host/db",
			expected: "postgres://user:[REDACTED]@host/db",
		},
		{
			name:     "fallback: invalid url encoding with complex password",
			input:    "postgres://user:p%ss:w@rd@host/db",
			expected: "postgres://user:[REDACTED]@host/db",
		},
		{
			name:     "fallback: query params with @ should not confuse",
			input:    "mysql://user:p%ssword@host/db?email=foo@bar.com",
			expected: "mysql://user:[REDACTED]@host/db?email=foo@bar.com",
		},
		{
			name:     "fallback: no password",
			input:    "user@host",
			expected: "user@host",
		},
		{
			name:     "fallback: no user info",
			input:    "host:port",
			expected: "host:port",
		},
		{
			name:     "fallback: no host separator",
			input:    "user:password",
			// If no host separator, we might fail to identify password.
			// Current regex expects @. Proposed logic expects @.
			// If no @, we assume no user/pass structure valid for DSN?
			// But "user:password" leaks password.
			// However, without @, it's hard to know if it's "host:port" or "user:pass".
			// We accept this limitation or try to guess.
			// The current regex returns original.
			expected: "user:password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactDSN(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
