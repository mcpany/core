package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRedactJSONFast_EscapedKeys(t *testing.T) {
	// "api_key" in hex: 61 70 69 5f 6b 65 79
	// "\u0061\u0070\u0069\u005f\u006b\u0065\u0079"

	input := `{"\u0061\u0070\u0069\u005f\u006b\u0065\u0079": "secret"}`
	// The default placeholder is "[REDACTED]" in this codebase (found in redact.go)
	expected := `{"\u0061\u0070\u0069\u005f\u006b\u0065\u0079": "[REDACTED]"}`

	got := redactJSONFast([]byte(input))
	assert.Equal(t, expected, string(got))
}

func TestRedactJSONFast_EscapedKeys_Mixed(t *testing.T) {
	// "api_key" mixed escapes
	// "a\u0070i_key"
	input := `{"a\u0070i_key": "secret"}`
	expected := `{"a\u0070i_key": "[REDACTED]"}`

	got := redactJSONFast([]byte(input))
	assert.Equal(t, expected, string(got))
}

func TestRedactJSONFast_AllEscapes(t *testing.T) {
	// Test all escape sequences handled by unescapeKeySmall
	// \b \f \n \r \t \" \\ \/
	// We construct a key that decodes to "api_key" but uses these escapes (if possible).
	// But "api_key" doesn't have backspace etc.
	// So we construct a key that HAS these chars, and ensure it is unescaped correctly.
	// But isKeySensitive only checks if it CONTAINS a sensitive word.
	// So "foo\bbar_api_key" -> unescapes to "foo<BS>bar_api_key". Matches "api_key".

	// \b \f \n \r \t \" \\ \/
	keyEscaped := `foo\b\f\n\r\t\"\\\/_api_key`
	input := `{"` + keyEscaped + `": "secret"}`
	expected := `{"` + keyEscaped + `": "[REDACTED]"}`

	got := redactJSONFast([]byte(input))
	assert.Equal(t, expected, string(got))
}


func TestRedactJSONFast_InvalidEscapes(t *testing.T) {
	// "api_key" but with invalid unicode escape
	// "api_key\uZZZZ" -> "api_keyu" -> not sensitive
	input := `{"api_key\uZZZZ": "secret"}`
	expected := `{"api_key\uZZZZ": "secret"}`

	got := redactJSONFast([]byte(input))
	assert.Equal(t, expected, string(got))
}

func TestRedactJSONFast_StackLimit(t *testing.T) {
	// Create a key slightly smaller than 256 bytes but with escapes
	// To trigger unescapeKeySmall

	// "api_key" padded with ignored chars? No, scanForSensitiveKeys matches any substring.
	// So "x...x_api_key" should match.

	prefix := ""
	for i := 0; i < 200; i++ {
		prefix += "x"
	}
	key := prefix + "_api_key" // 200 + 8 = 208 chars. < 256.

	input := `{"` + key + `": "secret"}`
	expected := `{"` + key + `": "[REDACTED]"}`

	got := redactJSONFast([]byte(input))
	assert.Equal(t, expected, string(got))
}
