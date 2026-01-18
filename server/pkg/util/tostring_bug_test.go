// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToString_Float32_LargeInteger(t *testing.T) {
	// 3 billion is larger than MaxInt32 (2.14 billion)
	// It is exactly representable in float32 because it has trailing zeros in binary.
	var val float32 = 3.0e9

	// We expect "3000000000", but current implementation returns "3e+09"
	expected := "3000000000"
	actual := ToString(val)

	assert.Equal(t, expected, actual, "float32 integer > MaxInt32 should be formatted as decimal integer")

	// Test boundary case: MaxInt32 + something small might round to same float32 if precision lost
	// But let's pick something that definitely is an integer in float32
	// 2^31 is 2147483648.
	val = 2147483648.0
	expected = "2147483648"
	actual = ToString(val)
	assert.Equal(t, expected, actual)

    // Test a negative large integer
    val = -3.0e9
    expected = "-3000000000"
    actual = ToString(val)
    assert.Equal(t, expected, actual)
}

func TestToString_Float32_Fractional(t *testing.T) {
    // Should still use normal formatting for non-integers
    var val float32 = 3.5
    expected := "3.5"
    actual := ToString(val)
    assert.Equal(t, expected, actual)
}
