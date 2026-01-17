// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_InvalidEscapes(t *testing.T) {
	// Case 1: Short key with invalid escape.
	// "password" escaped as "p\u0061ssword" (valid)
	// Add an invalid escape "\@" at the end.
	// JSON: {"p\u0061ssword\@": "secret"}
	// Expected: Redacted (because it contains "password")

	input := []byte(`{"p\u0061ssword\@": "secret_value"}`)
	expected := []byte(`{"p\u0061ssword\@": "[REDACTED]"}`)

	result := RedactJSON(input)
	assert.Equal(t, string(expected), string(result), "Short key with invalid escape should be redacted")
}

func TestRedactJSON_InvalidEscapes_LongKey(t *testing.T) {
	// Case 2: Long key with invalid escape.
	// We need to force using scanEscapedKeyForSensitive.

	oldLimit := maxUnescapeLimit
	maxUnescapeLimit = 100
	defer func() { maxUnescapeLimit = oldLimit }()

	// key length > 100.
	padding := strings.Repeat("a", 120)
	// "password" escaped
	escapedPassword := `p\u0061ssword`
	// Invalid escape
	invalidEscape := `\@`

	key := padding + escapedPassword + invalidEscape
	input := []byte(`{"` + key + `": "secret_value"}`)

	// Expected: Redacted
	expectedSuffix := `": "[REDACTED]"}`

	result := RedactJSON(input)

	assert.True(t, strings.HasSuffix(string(result), expectedSuffix), "Long key with invalid escape should be redacted")
}

func TestRedactJSON_InvalidEscapes_EdgeCases(t *testing.T) {
	// Test various invalid escapes
	cases := []struct {
		name string
		key  string
	}{
		{"Invalid hex", `p\u0061ssword\G`}, // \G is invalid escape, treated as 'G'. "passwordG" -> matched (uppercase G not skipped)
		{"Incomplete hex", `p\u0061ssword\1`}, // \1 is invalid (digits not valid escape start unless \u), treated as '1'. "password1" -> matched
		{"Unknown escape", `p\u0061ssword\?`}, // \? -> ? -> "password?" -> matched
		{"Trailing backslash", `p\u0061ssword\`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := []byte(`{"` + tc.key + `": "secret_value"}`)
			result := RedactJSON(input)

			// We check if it ends with "[REDACTED]"}`
			// Note: if the key has trailing backslash, the input string itself might be invalid JSON
			// because the quote is escaped? `{"key\": ...`
			// Let's check how we construct it.
			// `{"p\u0061ssword\": "secret_value"}` -> key is `p\u0061ssword"`, value is `secret_value`.
			// Wait, if we have `\` at the end of key content, it escapes the closing quote!
			// So `{"key\": "val"}` -> key is `key": "val`. It consumes the value?
			// `redactJSONFast` handles this by counting backslashes.

			if strings.HasSuffix(tc.key, `\`) {
				// If key ends with backslash, it escapes the quote.
				// So `{"key\": "val"}`
				// redactJSONFast will skip the quote after `key\` and look for the next one.
				// This might break the test assumption.
				// Let's ensure the key doesn't actually escape the quote for this test,
				// by escaping the backslash itself if we want to test literal backslash?
				// But we are testing "invalid escape sequence".
				// `\` at end of string IS valid if it escapes the quote.
				// If we mean "key ends with a backslash char", it must be `\\`.
				// If we mean "key contains \ followed by nothing (EOF)", that's impossible in a valid string.
				// The only way to have `\` that is NOT an escape for the following char is if it IS `\\`.

				// So "Trailing backslash" case is tricky in JSON.
				// `{"key\u0061\"`: unclosed.
				return
			}

			expectedSuffix := `": "[REDACTED]"}`
			if !strings.HasSuffix(string(result), expectedSuffix) {
				t.Errorf("Failed to redact key with %s. Result: %s", tc.name, string(result))
			}
		})
	}
}
