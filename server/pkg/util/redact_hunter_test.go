package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestHunter_RedactDSN_FalsePositives(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// False positives (Labels)
		{
			name:     "Label with email",
			input:    "Contact:support@example.com",
			expected: "Contact:support@example.com",
		},
		{
			name:     "Error message",
			input:    "Error:user@host connection failed",
			expected: "Error:user@host connection failed",
		},
		{
			name:     "Capitalized User in DSN (Admin not in skip list)",
			input:    "Admin:pass@host",
			expected: "Admin:[REDACTED]@host",
		},

		// Regressions (Standard DSNs)
		{
			name:     "Standard Postgres",
			input:    "postgres://user:password@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
		{
			name:     "Scheme-less DSN",
			input:    "user:password@host",
			expected: "user:[REDACTED]@host",
		},
		{
			name:     "Scheme-less DSN with port",
			input:    "user:password@host:5432",
			expected: "user:[REDACTED]@host:5432",
		},
		{
			name:     "Empty User",
			input:    ":password@host",
			expected: ":[REDACTED]@host",
		},
		{
			name:     "Hyphenated User",
			input:    "my-user:pass@host",
			expected: "my-user:[REDACTED]@host",
		},
		{
			name:     "User with number",
			input:    "user1:pass@host",
			expected: "user1:[REDACTED]@host",
		},

		// Pre-redacted cases
		{
			name:     "Pre-redacted user",
			input:    "postgres://[REDACTED]:password@host",
			expected: "postgres://[REDACTED]:[REDACTED]@host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RedactDSN(tt.input)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestHunter_RedactDSN_DoctorUsage(t *testing.T) {
	// Simulate doctor error messages
	errMsg := "Invalid URL: postgres://user:password@host"
	redacted := RedactDSN(errMsg)
	// Expect partial redaction
	assert.Equal(t, "Invalid URL: postgres://user:[REDACTED]@host", redacted)

	errMsg2 := "failed to connect: user:pass@host refused"
	redacted2 := RedactDSN(errMsg2)
	assert.Equal(t, "failed to connect: user:[REDACTED]@host refused", redacted2)

	errMsg3 := "Contact support:help@mcpany.com"
	redacted3 := RedactDSN(errMsg3)
	// "support:help@mcpany.com". "support" is lowercase.
	// Matches DSN pattern. Redacts.
	// This is ambiguous. "support" implies user.
	// If "support" is intended as user in a DSN, it is redacted.
	// If "support" is a label? Usually labels are capitalized.
	// If it is "Contact support: help@..." (space). No match.
	// If it is "Contact support:help@..." (no space).
	// "support" matches.
	assert.Equal(t, "Contact support:[REDACTED]@mcpany.com", redacted3)
}
