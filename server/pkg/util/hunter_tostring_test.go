// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToString_BugFix(t *testing.T) {
	// 2^63 is 9223372036854775808.
	// MaxInt64 is 9223372036854775807.
	// float64(MaxInt64) is 9223372036854775808 (rounded up).
	// Previously, ToString check val <= float64(math.MaxInt64) allowed 2^63.
	// casting 2^63 to int64 resulted in MinInt64 (overflow).

	val := math.Pow(2, 63)
	str := ToString(val)
	// It should NOT be negative. It should be scientific notation or a large positive number.
	assert.NotContains(t, str, "-", "ToString(2^63) should not be negative")
	// Updated requirement (2026): Large numbers (>= MaxInt64) should use scientific notation for safety.
	assert.Contains(t, str, "e+", "ToString(2^63) should be in scientific notation")
}

func TestToString_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "NaN",
			input:    math.NaN(),
			expected: "NaN",
		},
		{
			name:     "Inf",
			input:    math.Inf(1),
			expected: "+Inf",
		},
		{
			name:     "-Inf",
			input:    math.Inf(-1),
			expected: "-Inf",
		},
		{
			name:     "MinInt64 float",
			input:    float64(math.MinInt64),
			expected: "-9223372036854775808",
		},
		{
			name:     "MaxInt64 float - precision loss",
			// This value will be rounded to 2^63 when converted to float64, so it behaves like 2^63
			input:    float64(math.MaxInt64),
			expected: "9.223372036854776e+18",
		},
		{
			name:     "Largest safe float integer below MaxInt64",
			// 2^63 - 2048
			input:    math.Pow(2, 63) - 2048,
			expected: "9223372036854773760",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ToString(tt.input))
		})
	}
}
