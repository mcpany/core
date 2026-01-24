// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"math"
	"strconv"
	"testing"
)

func TestTokenCountFloatIntImplementation(t *testing.T) {
	// Test cases around powers of 10 and edges
	tests := []int64{
		1000000, 1000001, 1234567, 9999999,
		10000000, 10000001, 12300000,
		-1000000, -1234567,
		1230000,    // 1.23e+06
		1230000000, // 1.23e+09
		math.MinInt64,
	}

	// Add more random checks
	for i := 0; i < 100; i++ {
		tests = append(tests, int64(1000000+i*12345))
	}

	for _, n := range tests {
		// Optimization helper is only for |n| >= 1000000
		if n > -1000000 && n < 1000000 {
			continue
		}

		// Truth
		f := float64(n)
		s := strconv.FormatFloat(f, 'g', -1, 64)
		expectedLen := len(s)
		expectedTokens := expectedLen / 4
		if expectedTokens < 1 {
			expectedTokens = 1
		}

		// Verify the actual implementation
		gotTokens := tokenCountFloatInt(n)

		if gotTokens != expectedTokens {
			t.Errorf("Mismatch for %d: expected %d (len %d string '%s'), got %d", n, expectedTokens, expectedLen, s, gotTokens)
		}
	}
}
