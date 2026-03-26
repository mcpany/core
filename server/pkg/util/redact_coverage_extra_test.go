// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRedactJSON_Coverage ...
// Summary: TestRedactJSON_Coverage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// 1. Empty input
	assert.Equal(t, []byte(""), RedactJSON([]byte("")))

	// 2. Whitespace only
	assert.Equal(t, []byte("   "), RedactJSON([]byte("   ")))

	// 3. Comment only
	assert.Equal(t, []byte("// comment"), RedactJSON([]byte("// comment")))

	// 4. Non-object/array
	assert.Equal(t, []byte("123"), RedactJSON([]byte("123")))
	assert.Equal(t, []byte(`"string"`), RedactJSON([]byte(`"string"`)))

	// 5. Malformed JSON start
	assert.Equal(t, []byte("  x"), RedactJSON([]byte("  x")))
}

// TestIsKeyColon_Coverage ...
// Summary: TestIsKeyColon_Coverage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Key at end of input
	input := []byte(`"key"`)
	// endOffset is 5
	assert.False(t, isKeyColon(input, 5))

	// Key followed by space then EOF
	input2 := []byte(`"key"   `)
	assert.False(t, isKeyColon(input2, 5))
}
