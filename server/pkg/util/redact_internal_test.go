// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"testing"
)

func TestUnquoteBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal string",
			input:    `"hello world"`,
			expected: "hello world",
		},
		{
			name:     "Escaped quotes",
			input:    `"hello \"world\""`,
			expected: `hello "world"`,
		},
		{
			name:     "Escaped backslashes",
			input:    `"C:\\Windows\\System32"`,
			expected: `C:\Windows\System32`,
		},
		{
			name:     "Control characters",
			input:    `"line1\nline2\r\tindented"`,
			expected: "line1\nline2\r\tindented",
		},
		{
			name:     "Unicode simple",
			input:    `"hello \u0021"`,
			expected: "hello !",
		},
		{
			name:     "Unicode complex",
			input:    `"hello \u263a"`,
			expected: "hello ☺",
		},
		{
			name:     "Unicode multi-byte",
			input:    `"\u4e2d\u6587"`,
			expected: "中文",
		},
		{
			name:     "Invalid unicode - short",
			input:    `"test \u123"`,
			expected: `test \u123`,
		},
		{
			name:     "Invalid unicode - invalid hex",
			input:    `"test \u123z"`,
			expected: `test \u123z`,
		},
		{
			name:     "Trailing backslash",
			input:    `"test \"`,
			expected: `test \`,
		},
		{
			name:     "Unknown escape",
			input:    `"test \z"`,
			expected: `test \z`,
		},
		{
			name:     "Single quote (non-standard but supported)",
			input:    `"It\'s me"`,
			expected: `It's me`,
		},
		{
			name:     "Slash",
			input:    `"http:\/\/example.com"`,
			expected: `http://example.com`,
		},
		{
			name:     "Backspace and Formfeed",
			input:    `"a\bb\f"`,
			expected: "a\bb\f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unquoteBytes([]byte(tt.input))
			if string(result) != tt.expected {
				t.Errorf("unquoteBytes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRedactJSONFast_EdgeCases(t *testing.T) {
	// Tests specifically targeting logic branches in redactJSONFast
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Malformed JSON - unclosed string",
			input:    `{"key": "val`,
			expected: `{"key": "val`,
		},
		{
			name:     "Escaped sensitive key",
			input:    `{"\u0061pi_key": "secret"}`,
			expected: `{"\u0061pi_key": "[REDACTED]"}`,
		},
		{
			name:     "Escaped non-sensitive key",
			input:    `{"\u006bey": "value"}`,
			expected: `{"\u006bey": "value"}`,
		},
		{
			name:     "Key with only whitespace after",
			input:    `{"api_key"   : "secret"}`,
			expected: `{"api_key"   : "[REDACTED]"}`,
		},
		{
			name:     "Empty input",
			input:    ``,
			expected: ``,
		},
		{
			name:     "Just a string",
			input:    `"just a string"`,
			expected: `"just a string"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := redactJSONFast([]byte(tt.input))
			if !bytes.Equal(result, []byte(tt.expected)) {
				t.Errorf("redactJSONFast(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
