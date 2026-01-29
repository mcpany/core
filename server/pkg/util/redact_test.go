package util_test

import (
	"testing"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestRedactSecrets(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		secrets  []string
		expected string
	}{
		{
			name:     "no secrets",
			text:     "hello world",
			secrets:  nil,
			expected: "hello world",
		},
		{
			name:     "empty text",
			text:     "",
			secrets:  []string{"secret"},
			expected: "",
		},
		{
			name:     "single secret",
			text:     "this is a secret message",
			secrets:  []string{"secret"},
			expected: "this is a [REDACTED] message",
		},
		{
			name:     "multiple secrets",
			text:     "user: admin, pass: 12345",
			secrets:  []string{"admin", "12345"},
			expected: "user: [REDACTED], pass: [REDACTED]",
		},
		{
			name:     "overlapping secrets (substrings)",
			text:     "password is SuperSecretPassword",
			secrets:  []string{"SuperSecret", "SuperSecretPassword"},
			expected: "password is [REDACTED]", // Should match longest first
		},
		{
			name:     "empty secret in list",
			text:     "hello world",
			secrets:  []string{"hello", ""},
			expected: "[REDACTED] world", // Should ignore empty secret
		},
		{
			name:     "duplicate secrets",
			text:     "token: abc",
			secrets:  []string{"abc", "abc"},
			expected: "token: [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.RedactSecrets(tt.text, tt.secrets)
			assert.Equal(t, tt.expected, result)
		})
	}
}
