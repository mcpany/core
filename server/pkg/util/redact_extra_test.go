// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"testing"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestRedactMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:     "no sensitive keys",
			input:    map[string]interface{}{"name": "john", "age": 30},
			expected: map[string]interface{}{"name": "john", "age": 30},
		},
		{
			name:     "sensitive key (password)",
			input:    map[string]interface{}{"password": "secret", "name": "john"},
			expected: map[string]interface{}{"password": "[REDACTED]", "name": "john"},
		},
		{
			name:     "nested map sensitive",
			input:    map[string]interface{}{"user": map[string]interface{}{"password": "secret"}},
			expected: map[string]interface{}{"user": map[string]interface{}{"password": "[REDACTED]"}},
		},
		{
			name:     "slice of maps sensitive",
			input:    map[string]interface{}{"users": []interface{}{map[string]interface{}{"password": "s1"}, map[string]interface{}{"password": "s2"}}},
			expected: map[string]interface{}{"users": []interface{}{map[string]interface{}{"password": "[REDACTED]"}, map[string]interface{}{"password": "[REDACTED]"}}},
		},
		{
			name:     "nested slice of slice",
			input:    map[string]interface{}{"data": []interface{}{[]interface{}{map[string]interface{}{"token": "123"}}}},
			expected: map[string]interface{}{"data": []interface{}{[]interface{}{map[string]interface{}{"token": "[REDACTED]"}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.RedactMap(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "not json",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "simple json",
			input:    `{"name":"john"}`,
			expected: `{"name":"john"}`,
		},
		{
			name:     "sensitive json",
			input:    `{"password":"secret"}`,
			expected: `{"password":"[REDACTED]"}`,
		},
		{
			name:     "sensitive json whitespace",
			input:    ` { "password" : "secret" } `,
			expected: ` { "password" : "[REDACTED]" } `,
		},
		{
			name:     "array json",
			input:    `[{"password":"secret"}]`,
			expected: `[{"password":"[REDACTED]"}]`,
		},
		{
			name:     "nested object with comments",
			input:    `{"a": { /* comment */ "password": "secret" } }`,
			expected: `{"a": { /* comment */ "password": "[REDACTED]" } }`,
		},
		{
			name:     "array with comments",
			input:    `[ /* start */ {"password": "secret"} // end
]`,
			expected: `[ /* start */ {"password": "[REDACTED]"} // end
]`,
		},
		{
			name:     "escaped key sensitive",
			input:    `{"pass\u0077ord": "secret"}`,
			expected: `{"pass\u0077ord": "[REDACTED]"}`,
		},
		{
			name:     "malformed json string",
			input:    `{"password": "unc`,
			expected: `{"password": "[REDACTED]"`, // Best effort redaction even if malformed
		},
		{
			name:     "malformed json object",
			input:    `{"password": "secret"`,
			expected: `{"password": "[REDACTED]"`,
		},
		{
			name:     "slash but not comment",
			input:    `{"key": "a/b", "password": "secret"}`,
			expected: `{"key": "a/b", "password": "[REDACTED]"}`,
		},
		{
			name:     "escaped quote in key",
			input:    `{"pass\"word": "secret"}`,
			expected: `{"pass\"word": "secret"}`, // pass"word is not sensitive
		},
		{
			name:     "escaped backslash in key",
			input:    `{"pass\\word": "secret"}`,
			expected: `{"pass\\word": "secret"}`, // "pass\word" is not "password"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.RedactJSON([]byte(tt.input))
			assert.Equal(t, tt.expected, string(result))
		})
	}
}
