package util

import (
	"testing"
)

func TestRedactSensitiveKeysInJSONLike(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JSON string",
			input:    `{"password": "secret", "token": "abc"}`,
			expected: `{"password": "[REDACTED]", "token": "[REDACTED]"}`,
		},
		{
			name:     "Mixed string",
			input:    `Error: {"api_key": "12345"} failed`,
			expected: `Error: {"api_key": "[REDACTED]"} failed`,
		},
		{
			name:     "Case insensitive",
			input:    `{"Password": "Secret"}`,
			expected: `{"Password": "[REDACTED]"}`,
		},
		{
			name:     "No sensitive keys",
			input:    `{"user": "alice"}`,
			expected: `{"user": "alice"}`,
		},
		{
			name:     "With spaces",
			input:    `"secret" : "value"`,
			// Note: The regex replacement normalizes spaces around the colon to ": "
			// Wait, my replacement string was "\"$1\": \""+redactedPlaceholder+"\""
			// So it puts ": " (colon space).
			// Input had " : " (space colon space).
			// So it becomes "secret": "[REDACTED]" (no space before colon).
			expected: `"secret": "[REDACTED]"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactSensitiveKeysInJSONLike(tt.input)
			if got != tt.expected {
				t.Errorf("RedactSensitiveKeysInJSONLike() = %v, want %v", got, tt.expected)
			}
		})
	}
}
