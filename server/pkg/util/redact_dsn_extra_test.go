package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactDSN_Encoded(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		expected string
	}{
		{
			name:     "Standard URL Encoded Password",
			dsn:      "postgres://user:p%40ssword@localhost:5432/db",
			expected: "postgres://user:[REDACTED]@localhost:5432/db",
		},
		{
			name:     "Special Chars Encoded",
			dsn:      "mysql://user:foo%3Abar@localhost:3306/db",
			expected: "mysql://user:[REDACTED]@localhost:3306/db",
		},
		{
			name:     "Fallback Regex Encoded",
			dsn:      "scheme://user:p%40ssword@host",
			expected: "scheme://user:[REDACTED]@host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactDSN(tt.dsn)
			assert.Equal(t, tt.expected, got)
		})
	}
}
