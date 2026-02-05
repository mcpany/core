// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToString_Float64_LargeValue(t *testing.T) {
	// A value larger than int64 (> 9.22e18)
	// 1e20
	var val float64 = 1e20

	actual := ToString(val)
	// Updated requirement: Use standard Go 'g' behavior.
	// 1e20 in 'g' format is "1e+20"
	assert.Equal(t, "1e+20", actual, "Should format 1e20 using scientific notation (standard 'g' behavior)")
}

func TestToString_Float64_Boundary(t *testing.T) {
    // MaxInt64 is 9223372036854775807
    var val float64 = float64(math.MaxInt64)

    // Check behavior near boundary. Since math.MaxInt64 is exactly representable in float64
    // (after conversion which might lose some low-bit precision if it wasn't a power of 2,
    // but float64(int64(9223372036854775807)) is 9.223372036854776e+18 which is exactly representable)
    // Actually MaxInt64 is NOT exactly representable in float64. 2^63-1.
    // float64 has 53 bits.

    actual := ToString(val)
    // It should be either a decimal string representation of the casted value or scientific notation.
    // Given our updated logic, float64(math.MaxInt64) == 9223372036854775808 (2^63)
    // which exceeds math.MaxInt64, so it should use 'g' and return "9.223372036854776e+18"
    assert.Contains(t, actual, "e+")
}

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
