package tokenizer

import (
	"testing"
	"github.com/stretchr/testify/assert"
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
	// 10 items * 0.1 = 1.0 -> 1.
	inputTen := make([]int, 10)
	count, err = CountTokensInValue(wtSmall, inputTen)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// 15 items * 0.1 = 1.5 -> 1.
	// 20 items * 0.1 = 2.0 -> 2.
	inputTwenty := make([]int, 20)
	count, err = CountTokensInValue(wtSmall, inputTwenty)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}
