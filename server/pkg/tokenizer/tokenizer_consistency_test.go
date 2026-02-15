package tokenizer

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestHunter_WordTokenizer_Consistency(t *testing.T) {
	// Factor 1.9.
	// 1 word -> 1.9 -> 1 token.
	// 2 words -> 3.8 -> 3 tokens.
	wt := &WordTokenizer{Factor: 1.9}

	// Case 1: []int{1, 2}
	// Fast path treats this as 2 words -> 3 tokens.
	ints := []int{1, 2}
	countInts, err := CountTokensInValue(wt, ints)
	assert.NoError(t, err)

	// Case 2: []interface{}{1, 2}
	// Recursive path sums tokens: 1 + 1 = 2 tokens.
	ifaces := []interface{}{1, 2}
	countIfaces, err := CountTokensInValue(wt, ifaces)
	assert.NoError(t, err)

	if countInts != countIfaces {
		t.Errorf("Inconsistency detected! []int: %d, []interface{}: %d", countInts, countIfaces)
	}

    // Case 3: []string{"hello", "world"}
    // Fast path for []string sums tokens: 1 + 1 = 2 tokens.
    strs := []string{"hello", "world"}
    countStrs, err := CountTokensInValue(wt, strs)
    assert.NoError(t, err)

    // Case 4: Joined string "hello world"
    // 2 words -> 3 tokens.
    joined := "hello world"
    countJoined, err := CountTokensInValue(wt, joined)
    assert.NoError(t, err)

    if countStrs != countJoined {
        t.Errorf("Inconsistency detected! []string: %d, joined string: %d", countStrs, countJoined)
    }
}
