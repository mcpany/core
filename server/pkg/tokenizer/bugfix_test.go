// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestHunter_WordTokenizer_PrimitiveSlice_Consistent(t *testing.T) {
	// Factor 1.9. int(1.9) is 1.
	// For a slice of 10 ints, we expect consistency with []string or recursive traversal.
	// Each item is 1 word -> 1 * 1.9 = 1.9 -> 1 token.
	// 10 items -> 10 tokens.
	wt := &WordTokenizer{Factor: 1.9}

	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	count, err := CountTokensInValue(wt, input)
	if err != nil {
		t.Fatalf("CountTokensInValue failed: %v", err)
	}

	if count != 10 {
		t.Errorf("Got %d, expected 10 (consistent with per-item tokenization)", count)
	}
}

func TestHunter_WordTokenizer_PrimitiveSlice_Int64_Consistent(t *testing.T) {
    wt := &WordTokenizer{Factor: 1.5}
    input := []int64{1, 2, 3, 4}
    // Each item: 1 token.
    // Total 4.

    count, err := CountTokensInValue(wt, input)
    if err != nil {
        t.Fatalf("CountTokensInValue failed: %v", err)
    }

    if count != 4 {
        t.Errorf("Got %d, expected 4", count)
    }
}

func TestHunter_WordTokenizer_EdgeCases(t *testing.T) {
	// Test empty slice
	wt := &WordTokenizer{Factor: 1.5}
	inputEmpty := []int{}
	count, err := CountTokensInValue(wt, inputEmpty)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	// Test small factor < 1, resulting in count < 1 -> clamped to 1
	wtSmall := &WordTokenizer{Factor: 0.1}
	inputSmall := []int{1}
	// 1 * 0.1 = 0.1 -> 0. Clamped to 1.
	count, err = CountTokensInValue(wtSmall, inputSmall)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Test small factor, sufficient items
	// 10 items. Each item 1 token (clamped). Total 10.
	inputTen := make([]int, 10)
	count, err = CountTokensInValue(wtSmall, inputTen)
	assert.NoError(t, err)
	assert.Equal(t, 10, count)

	// 20 items. Each item 1 token (clamped). Total 20.
	inputTwenty := make([]int, 20)
	count, err = CountTokensInValue(wtSmall, inputTwenty)
	assert.NoError(t, err)
	assert.Equal(t, 20, count)
}
