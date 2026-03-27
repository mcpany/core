// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

func TestSkipString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    int
		expected int
	}{
		{
			name:     "simple string",
			input:    `"hello"`,
			start:    0,
			expected: 7, // index after last quote
		},
		{
			name:     "empty string",
			input:    `""`,
			start:    0,
			expected: 2,
		},
		{
			name:     "string with escaped quote",
			input:    `"hello\"world"`,
			start:    0,
			expected: 14,
		},
		{
			name:     "string with escaped backslash",
			input:    `"hello\\world"`,
			start:    0,
			expected: 14,
		},
		{
			name:     "string with escaped backslash at end",
			input:    `"val\\"`,
			start:    0,
			expected: 7, // "val\\" -> val\ (escaped backslash, then quote)
		},
		{
			name:     "string ending with escaped quote",
			input:    `"val\""`,
			start:    0,
			expected: 7, // "val\"" -> val" (escaped quote).
		},
		{
			name:     "string with double escaped backslash",
			input:    `"val\\\\"`,
			start:    0,
			expected: 9, // "val\\\\" -> val\\
		},
		{
			name:     "string with odd backslashes before quote (escaped quote)",
			input:    `"val\\\""`,
			start:    0,
			expected: 9, // "val\\\"" -> val\" (escaped quote). End of input.
		},
		{
			name:     "string with odd backslashes and closing quote",
			input:    `"val\\\""end"`,
			start:    0,
			expected: 9, // "val\\\"" -> val\" (escaped quote) then " (closing quote). "end" is outside.
		},
		{
			name:     "unclosed string",
			input:    `"hello`,
			start:    0,
			expected: 6,
		},
		{
			name:     "unclosed string with escaped quote",
			input:    `"hello\"`,
			start:    0,
			expected: 8,
		},
		{
			name:     "string not at start",
			input:    `key: "value"`,
			start:    5, // "value" starts at index 5
			expected: 12, // 5 + 7 = 12
		},
		{
			name:     "utf8 string",
			input:    `"hello world 🌍"`,
			start:    0,
			expected: 18,
		},
		{
			name:     "complex escapes",
			input:    `"foo\\\"bar"`,
			start:    0,
			expected: 12,
		},
		{
			name:     "many backslashes even",
			input:    `"\\\\\\\\"`, // 8 backslashes (4 real ones)
			start:    0,
			expected: 10,
		},
		{
			name:     "many backslashes odd",
			input:    `"\\\\\\\\\"`, // 9 backslashes (4 real ones + 1 escaping quote)
			start:    0,
			expected: 11, // Unclosed because quote is escaped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := skipString([]byte(tt.input), tt.start)
			if got != tt.expected {
				t.Errorf("skipString(%q, %d) = %d; want %d", tt.input, tt.start, got, tt.expected)
			}
		})
	}
}
