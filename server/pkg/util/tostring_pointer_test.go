// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

func TestToString_Pointer(t *testing.T) {
	str := "hello"
	ptrStr := &str

	val := 123
	ptrInt := &val

	boolean := true
	ptrBool := &boolean

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "pointer to string",
			input:    ptrStr,
			expected: "hello",
		},
		{
			name:     "pointer to int",
			input:    ptrInt,
			expected: "123",
		},
		{
			name:     "pointer to bool",
			input:    ptrBool,
			expected: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToString(tt.input)
			if got != tt.expected {
				// We expect this to fail initially (it will return address)
				// So we check if it starts with 0x to confirm it's failing in the way we expect
				// But strictly for the test, we want it to equal expected.
				t.Errorf("ToString() = %v, want %v", got, tt.expected)
			}
		})
	}
}
