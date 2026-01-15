// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRedactFast_Coverage(t *testing.T) {
	// Cover skipLiteral
	// true, false, null
	assert.Equal(t, []byte(`{"a": true, "b": false, "c": null}`), redactJSONFast([]byte(`{"a": true, "b": false, "c": null}`)))

	// Cover skipNumber
	assert.Equal(t, []byte(`{"a": 123}`), redactJSONFast([]byte(`{"a": 123}`)))
	assert.Equal(t, []byte(`{"a": 123.456}`), redactJSONFast([]byte(`{"a": 123.456}`)))

	// Cover skipObject nested with strings
	// Ensure we don't break on "}" inside string
	input := `{"a": {"b": "}"}}`
	assert.Equal(t, []byte(input), redactJSONFast([]byte(input)))

	// Cover skipArray
	inputArr := `{"a": ["b", "c"]}`
	assert.Equal(t, []byte(inputArr), redactJSONFast([]byte(inputArr)))

	// Cover sensitive key with array value
	// {"api_key": ["a", "b"]} -> {"api_key": "[REDACTED]"}
	inputSens := `{"api_key": ["a", "b"]}`
	expectedSens := `{"api_key": "[REDACTED]"}`
	assert.Equal(t, []byte(expectedSens), redactJSONFast([]byte(inputSens)))

	// Cover sensitive key with object value
	// {"api_key": {"a": "b"}} -> {"api_key": "[REDACTED]"}
	inputSensObj := `{"api_key": {"a": "b"}}`
	expectedSensObj := `{"api_key": "[REDACTED]"}`
	assert.Equal(t, []byte(expectedSensObj), redactJSONFast([]byte(inputSensObj)))

	// Incomplete value at EOF (sensitive)
	// {"api_key": -> {"api_key":
	// The redaction logic skips if value start is at EOF.
	assert.Equal(t, []byte(`{"api_key":`), redactJSONFast([]byte(`{"api_key":`)))

	// Incomplete literal at EOF
	assert.Equal(t, []byte(`{"a": true`), redactJSONFast([]byte(`{"a": true`)))
}

func TestScanEscapedKeyForSensitive_Coverage(t *testing.T) {
	// Lower the limit to force usage of scanEscapedKeyForSensitive
	oldLimit := maxUnescapeLimit
	maxUnescapeLimit = 5
	defer func() { maxUnescapeLimit = oldLimit }()

	tests := []struct {
		name     string
		key      string
		shouldRedact bool
	}{
		{
			name: "simple escape",
			key:  `\u0061uth_token`, // auth_token (len > 5)
			shouldRedact: true,
		},
		{
			name: "newline escape",
			key:  `auth\ntoken`,
			shouldRedact: true, // "token" is sensitive
		},
		{
			name: "tab escape",
			key:  `auth\ttoken`,
			shouldRedact: true,
		},
		{
			name: "other escapes",
			key:  `auth\b\f\rtoken`,
			shouldRedact: true,
		},
		{
			name: "quote escape",
			key:  `\"token\"`,
			shouldRedact: true,
		},
		{
			name: "backslash escape",
			key:  `\\token\\`,
			shouldRedact: true,
		},
		{
			name: "hex escape valid",
			key:  `\u0074oken`, // token
			shouldRedact: true,
		},
		{
			name: "hex escape partial",
			key:  `\u007`, // incomplete
			shouldRedact: false,
		},
		{
			name: "hex escape invalid char",
			key:  `\u007g`, // invalid hex
			shouldRedact: false, // u007g -> u matches nothing sensitive (token needs t)
		},
		{
			name: "hex escape incomplete at end",
			key:  `prefix\u00`,
			shouldRedact: false,
		},
		{
			name: "unknown escape",
			key:  `\ztoken`, // ztoken -> matches token? No, z is lowercase. ztoken != token.
			shouldRedact: true, // Wait. \z -> z. ztoken. "token" matches "ztoken"? No.
                                // But "token" matches "ztoken" at index 1.
                                // Boundary check: End of string.
                                // "token" in "ztoken". Yes.
		},
		{
			name: "split across buffer",
			// We need a very long string to trigger buffer refill
			// Buffer is 4096.
			// "token" is sensitive.
			// Fill 4094 chars then "token".
			key: strings.Repeat("a", 4094) + "token",
			shouldRedact: true,
		},
		{
			name: "split across buffer with escape",
			// Fill 4090 chars then \u0074oken
			key: strings.Repeat("a", 4090) + `\u0074oken`,
			shouldRedact: true,
		},
        {
            name: "escaped backslash at EOF",
            key: `token\\`,
            shouldRedact: true, // token matches
        },
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Construct JSON: {"<key>": "value"}
			input := `{"` + tc.key + `": "value"}`
			result := redactJSONFast([]byte(input))

			if tc.shouldRedact {
				assert.Contains(t, string(result), "[REDACTED]", "Key %q should be redacted", tc.key)
			} else {
				assert.NotContains(t, string(result), "[REDACTED]", "Key %q should NOT be redacted", tc.key)
			}
		})
	}
}
