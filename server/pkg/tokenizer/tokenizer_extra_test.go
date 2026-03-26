// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestCoverage_WordTokenizer_DeadCode ...
// Summary: TestCoverage_WordTokenizer_DeadCode
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	wt := NewWordTokenizer()
	visited := make(map[uintptr]bool)
	// Direct call to cover dead code in countRecursive
	c, err := wt.countRecursive(123, visited)
	assert.NoError(t, err)
	// 1 word * 1.3 = 1
	assert.Equal(t, 1, c)
}

// TestCoverage_WordTokenizer_CountTokens_SmallFactor ...
// Summary: TestCoverage_WordTokenizer_CountTokens_SmallFactor
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	wt := &WordTokenizer{Factor: 0.1}
	// "hello" -> 1 word. 1 * 0.1 = 0. Clamped to 1.
	c, err := wt.CountTokens("hello")
	assert.NoError(t, err)
	assert.Equal(t, 1, c)
}

// TestCoverage_CountTokensInValueWord_SmallFactor ...
// Summary: TestCoverage_CountTokensInValueWord_SmallFactor
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	wt := &WordTokenizer{Factor: 0.1}
	// 123 -> 1 word.
	c, err := CountTokensInValue(wt, 123)
	assert.NoError(t, err)
	assert.Equal(t, 1, c)
}
