// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_Comprehensive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// --- Basic Functionality ---
		{
			name:     "basic sensitive key",
			input:    `{"password": "secret"}`,
			expected: `{"password": "[REDACTED]"}`,
		},
		{
			name:     "basic non-sensitive key",
			input:    `{"username": "jdoe"}`,
			expected: `{"username": "jdoe"}`,
		},
		{
			name:     "multiple keys",
			input:    `{"username": "jdoe", "password": "secret"}`,
			expected: `{"username": "jdoe", "password": "[REDACTED]"}`,
		},

		// --- Comments Support ---
		{
			name:     "inline comment before colon",
			input:    `{"password" /* comment */ : "secret"}`,
			expected: `{"password" /* comment */ : "[REDACTED]"}`,
		},
		{
			name:     "inline comment after colon",
			input:    `{"password": /* comment */ "secret"}`,
			expected: `{"password": /* comment */ "[REDACTED]"}`,
		},
		{
			name:     "line comment before colon",
			input:    `{"password" // comment
: "secret"}`,
			expected: `{"password" // comment
: "[REDACTED]"}`,
		},
		{
			name:     "comment with structural chars",
			input:    `{"password" /* : } */ : "secret"}`,
			expected: `{"password" /* : } */ : "[REDACTED]"}`,
		},
		{
			name:     "comment inside non-sensitive value",
			input:    `{"user": ["a", /* comment */ "b"]}`,
			expected: `{"user": ["a", /* comment */ "b"]}`,
		},

		// --- Escaped Keys ---
		{
			name:     "escaped quotes in key",
			input:    `{"\"password\"": "secret"}`, // JSON key is "\"password\"" which unmarshals to "password"
			expected: `{"\"password\"": "[REDACTED]"}`,
		},
		{
			name:     "unicode escape in key (start)",
			input:    `{"\u0070assword": "secret"}`, // \u0070 is 'p'
			expected: `{"\u0070assword": "[REDACTED]"}`,
		},
		{
			name:     "unicode escape in key (middle)",
			input:    `{"pa\u0073sword": "secret"}`, // \u0073 is 's'
			expected: `{"pa\u0073sword": "[REDACTED]"}`,
		},
		{
			name:     "unicode escape in key (end)",
			input:    `{"passwor\u0064": "secret"}`, // \u0064 is 'd'
			expected: `{"passwor\u0064": "[REDACTED]"}`,
		},
		{
			name:     "mixed escapes",
			input:    `{"\u0070\u0061ssword": "secret"}`,
			expected: `{"\u0070\u0061ssword": "[REDACTED]"}`,
		},

		// --- Boundary & False Positives ---
		{
			name:     "prefix match (passport vs password)",
			input:    `{"passport": "valid"}`,
			expected: `{"passport": "valid"}`,
		},
		{
			name:     "prefix match (auth vs author)",
			input:    `{"author": "me"}`,
			expected: `{"author": "me"}`,
		},
		{
			name:     "prefix match (auth vs authority)",
			input:    `{"authority": "admin"}`,
			expected: `{"authority": "admin"}`,
		},
		{
			name:     "suffix match (my_password)",
			input:    `{"my_password": "secret"}`,
			expected: `{"my_password": "[REDACTED]"}`,
		},
		{
			name:     "camelCase match (authToken)",
			input:    `{"authToken": "secret"}`,
			expected: `{"authToken": "[REDACTED]"}`,
		},
		{
			name:     "camelCase no match (authorTool)",
			input:    `{"authorTool": "val"}`,
			expected: `{"authorTool": "val"}`,
		},

		// --- Complex Nesting & Types ---
		{
			name:     "object value redaction",
			input:    `{"password": {"k": "v", "a": [1, 2]}}`,
			expected: `{"password": "[REDACTED]"}`,
		},
		{
			name:     "array value redaction",
			input:    `{"secrets": ["s1", "s2"]}`,
			expected: `{"secrets": "[REDACTED]"}`,
		},
		{
			name:     "deep nesting",
			input:    `{"a": {"b": {"c": {"password": "deep"}}}}`,
			expected: `{"a": {"b": {"c": {"password": "[REDACTED]"}}}}`,
		},
		{
			name:     "sensitive key inside array of objects",
			input:    `{"users": [{"id": 1, "password": "s1"}, {"id": 2, "password": "s2"}]}`,
			expected: `{"users": [{"id": 1, "password": "[REDACTED]"}, {"id": 2, "password": "[REDACTED]"}]}`,
		},

		// --- Whitespace Variations ---
		{
			name:     "whitespace madness",
			input:    `{  "password"  :  "secret"  }`,
			expected: `{  "password"  :  "[REDACTED]"  }`,
		},
		{
			name:     "newlines and tabs",
			input:    "{\n\t\"password\":\n\t\"secret\"\n}",
			expected: "{\n\t\"password\":\n\t\"[REDACTED]\"\n}",
		},

		// --- Malformed/Resilience ---
		{
			name:     "unclosed object value",
			input:    `{"password": {"unclosed": 1`,
			expected: `{"password": "[REDACTED]"` , // Corrected: output does NOT contain closing brace because input didn't
		},
		{
			name:     "broken escape",
			input:    `{"password": "valid", "broken": "\\uZZZZ"}`,
			expected: `{"password": "[REDACTED]", "broken": "\\uZZZZ"}`,
		},

		// --- Unicode Values ---
		{
			name:     "unicode value preserved",
			input:    `{"user": "José", "password": "sí"}`,
			expected: `{"user": "José", "password": "[REDACTED]"}`,
		},

		// --- Values with delimiters ---
		{
			name:     "value with braces and brackets",
			input:    `{"public": "{ } [ ]", "password": "secret"}`,
			expected: `{"public": "{ } [ ]", "password": "[REDACTED]"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RedactJSON([]byte(tt.input))
			assert.Equal(t, tt.expected, string(output))
		})
	}
}
