package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesToString(t *testing.T) {
	b := []byte("hello")
	s := BytesToString(b)
	assert.Equal(t, "hello", s)
}

func TestRedactJSON_EscapeSequences(t *testing.T) {
	// Test various escape sequences in keys to hit unescapeKeySmall cases
	// We use "token" as the base sensitive key.
	// t = \u0074, o = \u006f, k = \u006b, e = \u0065, n = \u006e

	// \b backspace, \f formfeed, \n newline, \r return, \t tab
	// These are rarely used in keys but supported by JSON.
	// We construct a key that uses these but still matches "token" after unescaping?
	// No, "token" doesn't have these chars.
	// But unescapeKeySmall needs to handle them to correctly unescape ANY key.
	// If we have a key like "key\n" and we want to check if it is sensitive.
	// "key\n" is not sensitive.
	// But we want to ensure unescapeKeySmall runs and covers those lines.
	// It runs if key contains '\'.

	// So we create a JSON with keys using these escapes.
	// {"key\b": "val"}
	// redactJSONFast will call isKeySensitive("key\b").
	// isKeySensitive sees '\', calls unescapeKeySmall.
	// unescapeKeySmall hits case 'b'.

	input := []byte(`{"key\b": "val", "key\f": "val", "key\n": "val", "key\r": "val", "key\t": "val", "key\/": "val", "key\'": "val"}`)
	// None of these are sensitive, so no redaction.
	// But it executes the code.
	RedactJSON(input)

	// Now test unicode escapes
	// \u006b is 'k'. So "to\u006ben" is "token". Should be redacted.
	input2 := []byte(`{"to\u006ben": "secret"}`)
	expected2 := `{"to\u006ben": "[REDACTED]"}`
	result2 := RedactJSON(input2)
	assert.Equal(t, expected2, string(result2))

	// Test invalid unicode escapes to hit failure paths/fallback
	// \u12 (short) -> 'u'
	input3 := []byte(`{"to\u12": "val"}`) // "to\u12" -> "tou12" (invalid hex handling in unescapeKeySmall loops)
	RedactJSON(input3)

	// \u123Z (invalid char)
	input4 := []byte(`{"to\u123Z": "val"}`)
	RedactJSON(input4)
}

func TestRedactJSON_BufferLimit(t *testing.T) {
	// Trigger buffer limit in unescapeKeySmall
	// unescapeStackLimit is 256.
	// We need a key > 256 bytes with escapes.
	// But < maxUnescapeLimit (1MB).

	// Generate key of 300 chars with a backslash
	key := ""
	for i := 0; i < 300; i++ {
		key += "a"
	}
	key += "\\u0061" // adds 'a'

	input := []byte(`{"` + key + `": "val"}`)
	RedactJSON(input)
}
