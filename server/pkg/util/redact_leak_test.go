package util

import (
	"testing"
)

func TestRedactDSN_Leak_QueryParameters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Leak: Postgres DSN with password in query",
			input:    "postgres://user:pass@host/db?password=supersecret&sslmode=disable",
			expected: "postgres://user:[REDACTED]@host/db?password=[REDACTED]&sslmode=disable",
		},
		{
			name:     "Leak: MySQL DSN with secret in query",
			input:    "mysql://user@tcp(localhost)/dbname?secret=hidden_value",
			expected: "mysql://user@tcp(localhost)/dbname?secret=[REDACTED]",
		},
		{
			name:     "Leak: Token in query",
			input:    "http://example.com/api?token=abcdef123456",
			expected: "http://example.com/api?token=[REDACTED]",
		},
		{
			name:     "Leak: Multiple sensitive params",
			input:    "https://api.example.com?api_key=123&other=safe&password=secret",
			expected: "https://api.example.com?api_key=[REDACTED]&other=safe&password=[REDACTED]",
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
