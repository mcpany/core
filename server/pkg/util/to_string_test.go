// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"testing"
)

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "json.Number",
			input:    json.Number("123.45"),
			expected: "123.45",
		},
		{
			name:     "bool true",
			input:    true,
			expected: "true",
		},
		{
			name:     "bool false",
			input:    false,
			expected: "false",
		},
		{
			name:     "int",
			input:    123,
			expected: "123",
		},
		{
			name:     "int64",
			input:    int64(1234567890),
			expected: "1234567890",
		},
		{
			name:     "float64",
			input:    123.456,
			expected: "123.456",
		},
		{
			name:     "fmt.Stringer",
			input:    testStringer{val: "test"},
			expected: "test",
		},
		{
			name:     "nil",
			input:    nil,
			expected: "<nil>",
		},
		{
			name:     "struct",
			input:    struct{ A int }{A: 1},
			expected: "{1}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToString(tt.input)
			if result != tt.expected {
				t.Errorf("ToString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

type testStringer struct {
	val string
}

func (ts testStringer) String() string {
	return ts.val
}
