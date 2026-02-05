// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRedactJSONFast_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic Redaction",
			input:    `{"token": "123"}`,
			expected: `{"token": "[REDACTED]"}`,
		},
		{
			name:     "Malformed JSON - Unclosed String",
			input:    `{"token": "123`,
			expected: `{"token": "[REDACTED]"`,
		},
		{
			name:     "Malformed JSON - Missing Value",
			input:    `{"token": }`,
			expected: `{"token": "[REDACTED]"}`, // Current behavior, arguably "fixing" it or making it valid string
		},
		{
			name:     "Escaped Key",
			input:    `{"\u0074oken": "123"}`,
			expected: `{"\u0074oken": "[REDACTED]"}`,
		},
		{
			name:     "False Positive Prefix",
			input:    `{"tokenizer": "123"}`,
			expected: `{"tokenizer": "123"}`,
		},
		{
			name:     "False Positive Suffix",
			input:    `{"broken": "123"}`,
			expected: `{"broken": "123"}`,
		},
		{
			name:     "Boundary CamelCase",
			input:    `{"authToken": "123"}`,
			expected: `{"authToken": "[REDACTED]"}`,
		},
		{
			name:     "Boundary PascalCase",
			input:    `{"AuthToken": "123"}`,
			expected: `{"AuthToken": "[REDACTED]"}`,
		},
		{
			name:     "Non-Sensitive PascalCase",
			input:    `{"Authority": "123"}`,
			expected: `{"Authority": "123"}`,
		},
		{
			name:     "Uppercase Key",
			input:    `{"TOKEN": "123"}`,
			expected: `{"TOKEN": "[REDACTED]"}`,
		},
		{
			name:     "Uppercase Non-Sensitive",
			input:    `{"AUTHOR": "123"}`,
			expected: `{"AUTHOR": "123"}`,
		},
		{
			name:     "Nested Map",
			input:    `{"a": {"token": "123"}}`,
			expected: `{"a": {"token": "[REDACTED]"}}`,
		},
		{
			name:     "Array of Objects",
			input:    `[{"token": "123"}]`,
			expected: `[{"token": "[REDACTED]"}]`,
		},
		{
			name:     "Key with null byte",
			input:    "{\"token\x00\": \"123\"}",
			expected: "{\"token\x00\": \"[REDACTED]\"}",
		},
		{
			name:     "Key with weird escape",
			input:    `{"token\n": "123"}`,
			expected: `{"token\n": "[REDACTED]"}`,
		},
		{
			name:     "Value with comment style chars",
			input:    `{"token": // comment}`,
			expected: `{"token": // comment}`, // Value is missing, so nothing is redacted
		},
		{
			name:     "Comment before value",
			input:    `{"token": // comment
 "secret"}`,
			expected: `{"token": // comment
 "[REDACTED]"}`,
		},
		{
			name:     "Block comment before value",
			input:    `{"token": /* comment */ "secret"}`,
			expected: `{"token": /* comment */ "[REDACTED]"}`,
		},
		{
			name:     "Value is valid number",
			input:    `{"token": 123}`,
			expected: `{"token": "[REDACTED]"}`,
		},
		{
			name:     "Value is float",
			input:    `{"token": 123.456}`,
			expected: `{"token": "[REDACTED]"}`,
		},
		{
			name:     "Value is scientific",
			input:    `{"token": 1e10}`,
			expected: `{"token": "[REDACTED]"}`,
		},
		{
			name:     "Value is true",
			input:    `{"token": true}`,
			expected: `{"token": "[REDACTED]"}`,
		},
		{
			name:     "Value is null",
			input:    `{"token": null}`,
			expected: `{"token": "[REDACTED]"}`,
		},
		// Escape coverage
		{
			name:     "Key with backspace escape",
			input:    `{"tok\ben": "123"}`, // tok<BS>en -> token (if BS ignored?) No, \b is byte 8.
			// unescapeKeySmall converts \b to byte 8.
			// scanForSensitiveKeys treats it as byte 8.
			// token matches token.
			// if key is "tok\ben". unescaped: 't','o','k',8,'e','n'.
			// token does not match.
			expected: `{"tok\ben": "123"}`,
		},
		{
			name:     "Key with unknown escape",
			input:    `{"tok\zen": "123"}`, // \z -> z
			expected: `{"tok\zen": "123"}`, // tokzen -> matches token? No.
		},
		{
			name:     "Key with unknown escape making sensitive",
			input:    `{"to\ken": "123"}`, // \k -> k. token.
			expected: `{"to\ken": "[REDACTED]"}`,
		},
		{
			name:     "Key with unicode escape invalid",
			input:    `{"to\u00ZZken": "123"}`, // \u00ZZ -> u00ZZ. tou00ZZken.
			expected: `{"to\u00ZZken": "123"}`,
		},
        // Edge case: Large key
        {
            name: "Large Key Redaction",
            input: `{"` + makeLargeString(200) + `token": "123"}`,
            expected: `{"` + makeLargeString(200) + `token": "[REDACTED]"}`,
        },
        // Edge case: Key split across buffer (simulated by logic?)
        // Hard to simulate buffer split in unit test without mocking.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := RedactJSON([]byte(tc.input))
			assert.Equal(t, tc.expected, string(got))
		})
	}
}

func makeLargeString(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = 'a'
    }
    return string(b)
}

func TestBytesToString(t *testing.T) {
	b := []byte("hello")
	s := BytesToString(b)
	assert.Equal(t, "hello", s)
}
