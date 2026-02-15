package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_EscapedKeys_Coverage(t *testing.T) {
	// Test escaping to match sensitive keys
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "unicode escape",
			input:    `{"pass\u0077ord": "secret"}`,
			expected: `{"pass\u0077ord": "[REDACTED]"}`,
		},
		{
			name:     "unknown escape treated as char",
			input:    `{"pass\word": "secret"}`, // \w -> w
			expected: `{"pass\word": "[REDACTED]"}`,
		},
		{
			name:     "invalid unicode escape treated as u",
			input:    `{"pass\uZZZZord": "val"}`, // -> passuZZZZord (not sensitive)
			expected: `{"pass\uZZZZord": "val"}`,
		},
        {
            name:     "short unicode escape",
            input:    `{"pass\u12": "val"}`, // -> passu12 (not sensitive)
            expected: `{"pass\u12": "val"}`,
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := RedactJSON([]byte(tt.input))
			assert.Equal(t, tt.expected, string(actual))
		})
	}
}

func TestRedactJSON_LargeKey_Coverage(t *testing.T) {
    // Test large key to force fallback or streaming path
    // 256 is unescapeStackLimit
    // We need a key > 256 chars that contains escapes.

    longKey := "x"
    for i := 0; i < 300; i++ {
        longKey += "x"
    }
    longKey += "password"

    // With escape to force isKeySensitive to go to unescape path
    escapedKey := strings.Replace(longKey, "password", "pass\\u0077ord", 1)

    input := `{"` + escapedKey + `": "secret"}`
    expected := `{"` + escapedKey + `": "[REDACTED]"}`

    actual := RedactJSON([]byte(input))
    assert.Equal(t, expected, string(actual))
}
