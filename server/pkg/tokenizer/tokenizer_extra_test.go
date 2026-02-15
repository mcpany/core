package tokenizer

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCoverage_WordTokenizer_DeadCode(t *testing.T) {
	wt := NewWordTokenizer()
	visited := make(map[uintptr]bool)
	// Direct call to cover dead code in countRecursive
	c, err := wt.countRecursive(123, visited)
	assert.NoError(t, err)
	// 1 word * 1.3 = 1
	assert.Equal(t, 1, c)
}

func TestCoverage_WordTokenizer_CountTokens_SmallFactor(t *testing.T) {
	wt := &WordTokenizer{Factor: 0.1}
	// "hello" -> 1 word. 1 * 0.1 = 0. Clamped to 1.
	c, err := wt.CountTokens("hello")
	assert.NoError(t, err)
	assert.Equal(t, 1, c)
}

func TestCoverage_CountTokensInValueWord_SmallFactor(t *testing.T) {
	wt := &WordTokenizer{Factor: 0.1}
	// 123 -> 1 word.
	c, err := CountTokensInValue(wt, 123)
	assert.NoError(t, err)
	assert.Equal(t, 1, c)
}
