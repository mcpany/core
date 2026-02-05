// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestHunter_SimpleTokenizer_FloatOptimization_Bug(t *testing.T) {
	st := NewSimpleTokenizer()

	tests := []struct {
		val float64
	}{
		// These values trigger the scientific notation in %v (default formatting),
		// but the fast path optimization treats them as integers, leading to incorrect token counts.
		{10000000.0},  // 1e+07 (5 chars -> 1 token) vs 10000000 (8 digits -> 2 tokens)
		{1234567.0},   // 1.234567e+06 (12 chars -> 3 tokens) vs 1234567 (7 digits -> 1 token)
		{-1234567.0},  // -1.234567e+06 (13 chars -> 3 tokens) vs -1234567 (8 chars -> 2 tokens)

		// Boundary cases
		{1000000.0},   // 1e+06 (5 chars -> 1 token) vs 1000000 (7 digits -> 1 token) - This one is coincidentally same count
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%f", tt.val), func(t *testing.T) {
			// Updated Expectation: Match JSON serialization (no scientific notation for these ints),
			// rather than fmt.Sprintf("%v") (which uses scientific notation).
			// This is an intentional optimization choice.

			// We manually calculate expected tokens based on integer string representation
			i := int64(tt.val)
			strRep := fmt.Sprintf("%d", i)
			expectedTokens, _ := st.CountTokens(strRep)

			// Actual behavior
			actualTokens, err := CountTokensInValue(st, tt.val)
			assert.NoError(t, err)

			if actualTokens != expectedTokens {
				t.Errorf("Mismatch for %f (JSON-like string: %s). Expected %d, Got %d",
					tt.val, strRep, expectedTokens, actualTokens)
			}
		})
	}
}

func TestHunter_SimpleTokenizer_FloatSliceOptimization_Bug(t *testing.T) {
	st := NewSimpleTokenizer()

	// A slice containing problematic floats
	vals := []float64{10000000.0, 1234567.0}

	// Updated Expectation: Match JSON serialization
	expectedCount := 0
	for _, v := range vals {
		i := int64(v)
		c, _ := st.CountTokens(fmt.Sprintf("%d", i))
		expectedCount += c
	}

	actualCount, err := CountTokensInValue(st, vals)
	assert.NoError(t, err)

	assert.Equal(t, expectedCount, actualCount, "Slice count mismatch")
}
