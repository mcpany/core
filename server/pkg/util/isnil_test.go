// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestIsNil(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "nil interface",
			input:    nil,
			expected: true,
		},
		{
			name:     "nil pointer",
			input:    (*int)(nil),
			expected: true,
		},
		{
			name:     "non-nil pointer",
			input:    new(int),
			expected: false,
		},
		{
			name:     "nil map",
			input:    (map[string]int)(nil),
			expected: true,
		},
		{
			name:     "non-nil map",
			input:    make(map[string]int),
			expected: false,
		},
		{
			name:     "nil slice",
			input:    ([]int)(nil),
			expected: true,
		},
		{
			name:     "non-nil slice",
			input:    make([]int, 0),
			expected: false,
		},
		{
			name:     "nil chan",
			input:    (chan int)(nil),
			expected: true,
		},
		{
			name:     "non-nil chan",
			input:    make(chan int),
			expected: false,
		},
		{
			name:     "nil func",
			input:    (func())(nil),
			expected: true,
		},
		{
			name:     "non-nil func",
			input:    func() {},
			expected: false,
		},
		{
			name:     "struct value (not nil)",
			input:    struct{}{},
			expected: false,
		},
		{
			name:     "int value (not nil)",
			input:    123,
			expected: false,
		},
		{
			name:     "string value (not nil)",
			input:    "hello",
			expected: false,
		},
		{
			name:     "nil interface inside interface",
			input:    func() interface{} { var i interface{}; return i }(),
			expected: true,
		},
		{
			name:     "array value (not nil)",
			input:    [3]int{1, 2, 3},
			expected: false,
		},
		{
			name:     "nil unsafe.Pointer",
			input:    (unsafe.Pointer)(nil),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsNil(tt.input))
		})
	}
}
