// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"math"
	"testing"
)

func TestToString_Float64Overflow(t *testing.T) {
	// math.MaxInt64 is 9223372036854775807
	// float64(math.MaxInt64) is 9223372036854775808 (2^63)
	// Casting this float back to int64 overflows to -9223372036854775808 (MinInt64)

	val := float64(math.MaxInt64)
	result := ToString(val)

	// We expect scientific notation or at least a positive number.
	// Definitely NOT starting with "-"
	if len(result) > 0 && result[0] == '-' {
		t.Errorf("ToString(float64(MaxInt64)) returned negative number: %s. Expected positive.", result)
	}

	// Check specifically that it doesn't return the MinInt64 string representation
	if result == "-9223372036854775808" {
		t.Errorf("ToString(float64(MaxInt64)) incorrectly wrapped around to MinInt64")
	}
}
