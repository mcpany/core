// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

func TestWalkJSONStrings_Bug_Comments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "string in block comment",
			input:    `/* "hidden" */ "visible"`,
			expected: `/* "hidden" */ "REPLACED"`,
		},
		{
			name:     "string in line comment",
			input:    `// "hidden"
"visible"`,
			expected: `// "hidden"
"REPLACED"`,
		},
		{
			name:     "string in block comment before key",
			input:    `{ /* "ignore" */ "key": "value" }`,
			expected: `{ /* "ignore" */ "key": "REPLACED" }`,
		},
		{
			name:     "slash not comment",
			input:    `/ "visible"`,
			expected: `/ "REPLACED"`,
		},
		{
			name:     "string in comment crossing quote boundary",
			input:    `/* "oops */ "real"`,
			expected: `/* "oops */ "REPLACED"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := WalkJSONStrings([]byte(tt.input), func(raw []byte) ([]byte, bool) {
				// Replace everything with "REPLACED"
				return []byte(`"REPLACED"`), true
			})

			if string(output) != tt.expected {
				t.Errorf("WalkJSONStrings() = %s, want %s", output, tt.expected)
			}
		})
	}
}
