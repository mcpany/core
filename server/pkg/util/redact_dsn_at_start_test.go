package util

import (
	"testing"
)

func TestRedactDSN_PasswordWithAtStart(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Password starts with @",
			input:    "user:@ssword@host",
			expected: "user:[REDACTED]@host",
		},
		{
			name:     "Password is just @",
			input:    "user:@@host",
			expected: "user:[REDACTED]@host",
		},
		{
			name:     "Schemeless with password starting with @",
			input:    "redis://user:@ssword@host",
			expected: "redis://user:[REDACTED]@host",
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
