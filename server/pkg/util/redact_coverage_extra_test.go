package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_Coverage(t *testing.T) {
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

func TestIsKeyColon_Coverage(t *testing.T) {
	// Key at end of input
	input := []byte(`"key"`)
	// endOffset is 5
	assert.False(t, isKeyColon(input, 5))

	// Key followed by space then EOF
	input2 := []byte(`"key"   `)
	assert.False(t, isKeyColon(input2, 5))
}
