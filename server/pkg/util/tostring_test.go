// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
    "math"
    "strconv"
)

func TestToString_LargeInteger(t *testing.T) {
	// 2^63 = 9223372036854775808
    // This value is exactly representable in float64
	val := float64(9223372036854775808.0)
	str := ToString(val)

	// We expect full digits, not scientific notation
	expected := "9223372036854775808"
	if str != expected {
		t.Errorf("ToString(%f) = %q, want %q", val, str, expected)
	}

    // Test a value that is NOT an integer
    valFloat := 1.23
    strFloat := ToString(valFloat)
    expectedFloat := "1.23"
    if strFloat != expectedFloat {
        t.Errorf("ToString(%f) = %q, want %q", valFloat, strFloat, expectedFloat)
    }

    // Test a very large integer (1e20)
    // 1e20 = 100000000000000000000
    valLarge := 1e20
    strLarge := ToString(valLarge)
    expectedLarge := "100000000000000000000"
    if strLarge != expectedLarge {
         t.Errorf("ToString(%f) = %q, want %q", valLarge, strLarge, expectedLarge)
    }
}

func TestToString_MaxInt64Boundaries(t *testing.T) {
    // MaxInt64 = 9223372036854775807
    maxInt := int64(math.MaxInt64)
    // float64(maxInt) rounds to 9223372036854775808 (2^63) because of precision loss
    val := float64(maxInt)

    // We expect the string representation of the float value
    expected := "9223372036854775808"

    if ToString(val) != expected {
        t.Errorf("ToString(float64(MaxInt64)) = %q, want %q", ToString(val), expected)
    }

    // MinInt64 = -9223372036854775808
    minInt := int64(math.MinInt64)
    // float64(minInt) is exactly -9223372036854775808 (2^63 is exact)
    valMin := float64(minInt)

    expectedMin := strconv.FormatInt(minInt, 10)

    if ToString(valMin) != expectedMin {
        t.Errorf("ToString(float64(MinInt64)) = %q, want %q", ToString(valMin), expectedMin)
    }
}
