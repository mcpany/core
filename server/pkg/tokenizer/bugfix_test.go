// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"testing"
)

func TestHunter_WordTokenizer_PrimitiveSlice_Undercount(t *testing.T) {
	// Factor 1.9. int(1.9) is 1.
	// For a slice of 10 ints, fast path returns 10 * 1 = 10.
	// We expect 10 * 1.9 = 19.
	wt := &WordTokenizer{Factor: 1.9}

	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	count, err := CountTokensInValue(wt, input)
	if err != nil {
		t.Fatalf("CountTokensInValue failed: %v", err)
	}

	// We expect roughly 19.
	// Current implementation gives 10.
	if count < 15 {
		t.Errorf("WordTokenizer undercounts primitive slice. Got %d, expected roughly 19 (Factor 1.9, 10 items)", count)
	}
}

func TestHunter_WordTokenizer_PrimitiveSlice_Int64(t *testing.T) {
    wt := &WordTokenizer{Factor: 1.5}
    input := []int64{1, 2, 3, 4}
    // Expected: 4 * 1.5 = 6.
    // Current: 4 * 1 = 4.

    count, err := CountTokensInValue(wt, input)
    if err != nil {
        t.Fatalf("CountTokensInValue failed: %v", err)
    }

    if count != 6 {
        t.Errorf("WordTokenizer undercounts int64 slice. Got %d, expected 6", count)
    }
}
