// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

func TestWalkJSONStrings_BugReproduction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		visitor  func(raw []byte) ([]byte, bool)
		expected string
	}{
		{
			name:  "bug reproduction: string in comment with preceding slash",
			input: `/ // "hidden"`,
			visitor: func(raw []byte) ([]byte, bool) {
				return []byte(`"REPLACED"`), true
			},
			expected: `/ // "hidden"`,
		},
		{
			name:  "bug reproduction: string in comment with preceding slash and spaces",
			input: `  /  // "hidden"`,
			visitor: func(raw []byte) ([]byte, bool) {
				return []byte(`"REPLACED"`), true
			},
			expected: `  /  // "hidden"`,
		},
		{
			name:  "bug reproduction: multiple slashes",
			input: `/ / / // "hidden"`,
			visitor: func(raw []byte) ([]byte, bool) {
				return []byte(`"REPLACED"`), true
			},
			expected: `/ / / // "hidden"`,
		},
		{
			name:  "bug reproduction: block comment with preceding slash",
			input: `/ /* "hidden" */`,
			visitor: func(raw []byte) ([]byte, bool) {
				return []byte(`"REPLACED"`), true
			},
			expected: `/ /* "hidden" */`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WalkJSONStrings([]byte(tt.input), tt.visitor)
			if string(result) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(result))
			}
		})
	}
}
