package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_Fast_Extended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "object with braces in string",
			input:    `{"public": "value with { braces }", "api_key": "secret"}`,
			expected: `{"public": "value with { braces }", "api_key": "[REDACTED]"}`,
		},
		{
			name:     "array with brackets in string",
			input:    `{"public": ["val [ bracket ]"], "api_key": "secret"}`,
			expected: `{"public": ["val [ bracket ]"], "api_key": "[REDACTED]"}`,
		},
		{
			name:     "unclosed object",
			input:    `{"api_key": {"unclosed": "object"`,
			expected: `{"api_key": "[REDACTED]"` ,
		},
		{
			name:     "unclosed array",
			input:    `{"api_key": ["unclosed", "array"`,
			expected: `{"api_key": "[REDACTED]"` ,
		},
		{
			name:     "literal true",
			input:    `{"api_key": true}`,
			expected: `{"api_key": "[REDACTED]"}`,
		},
		{
			name:     "literal false",
			input:    `{"api_key": false}`,
			expected: `{"api_key": "[REDACTED]"}`,
		},
		{
			name:     "literal null",
			input:    `{"api_key": null}`,
			expected: `{"api_key": "[REDACTED]"}`,
		},
		{
			name:     "number integer",
			input:    `{"api_key": 12345}`,
			expected: `{"api_key": "[REDACTED]"}`,
		},
		{
			name:     "number float",
			input:    `{"api_key": 123.456}`,
			expected: `{"api_key": "[REDACTED]"}`,
		},
		{
			name:     "number negative",
			input:    `{"api_key": -123}`,
			expected: `{"api_key": "[REDACTED]"}`,
		},
		{
			name:     "number exponent",
			input:    `{"api_key": 1.23e5}`,
			expected: `{"api_key": "[REDACTED]"}`,
		},
		{
			name:     "malformed string (unclosed)",
			input:    `{"api_key": "unclosed string`,
			expected: `{"api_key": "[REDACTED]"` ,
		},
		{
			name:     "malformed string (odd backslashes at EOF)",
			input:    `{"api_key": "ends with slash \`,
			expected: `{"api_key": "[REDACTED]"` ,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RedactJSON([]byte(tt.input))
			assert.Equal(t, tt.expected, string(output))
		})
	}
}

func TestRedactJSON_KeyCheck(t *testing.T) {
	// Cover the `if malformed { break }` logic in key parsing
	t.Run("malformed key", func(t *testing.T) {
		input := `{"unclosed key`
		output := RedactJSON([]byte(input))
		assert.Equal(t, input, string(output))
	})

	// Cover escaped key check logic with length limit
	t.Run("large escaped key", func(t *testing.T) {
		// Key larger than 1024 bytes
		longKey := "pass" + "\\u0077" + "ord" + string(make([]byte, 2000))
		input := `{"` + longKey + `": "val"}`
		// It should NOT unmarshal (too large), but use scanEscapedKeyForSensitive.
		// scanEscapedKeyForSensitive handles `\u0077` -> 'w'.
		// So it matches "password".
		// Previous behavior (bug): fell back to raw key which didn't match.
		// Fixed behavior: matches and redacts.
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
	})

	// Cover `skipJSONValue` with unknown char
	t.Run("unknown value type", func(t *testing.T) {
		// e.g. starting with 'x' (invalid JSON)
		input := `{"api_key": xxxxx}`
		// skipJSONValue calls skipNumber by default which scans until delimiter.
		// xxxxx ends at }
		// So it should be redacted.
		output := RedactJSON([]byte(input))
		expected := `{"api_key": "[REDACTED]"}`
		assert.Equal(t, expected, string(output))
	})
}
