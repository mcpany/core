package util

import (
	"testing"
)

func TestRedactDSN_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Password with colon",
			input:    "postgres://user:pass:word@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
		{
			name:     "Password with at sign",
			input:    "postgres://user:pass@word@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
		{
			name: "Malformed URL that url.Parse fails",
			input: "not a url user:password@host",
			expected: "not a url user:[REDACTED]@host",
		},
		{
			name: "Invalid escape sequence with colon in password",
			input: "postgres://user:pass:word%@host",
			expected: "postgres://user:[REDACTED]@host",
		},
		{
			name: "Invalid escape sequence with at sign in password",
			input: "postgres://user:pass@word%@host",
			expected: "postgres://user:[REDACTED]@host",
		},
		{
			name:     "Empty username",
			input:    "postgres://:password@host",
			expected: "postgres://:[REDACTED]@host",
		},
		{
			name:     "Empty username with colon in password",
			input:    "postgres://:pass:word@host",
			expected: "postgres://:[REDACTED]@host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactDSN(tt.input)
			if got != tt.expected {
				t.Errorf("RedactDSN(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
