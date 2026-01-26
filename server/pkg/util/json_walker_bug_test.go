package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalkJSONStrings_Bug_SlashBeforeComment(t *testing.T) {
	// Input contains a division operator (slash) followed by a line comment containing a string.
	// The walker should skip the comment and NOT visit "hidden".
	input := []byte(`{ "val": 1 / 1 // "hidden" }`)

	visited := []string{}
	WalkJSONStrings(input, func(raw []byte) ([]byte, bool) {
		visited = append(visited, string(raw))
		return nil, false
	})

	// We expect NO values visited.
	// "val" is a key.
	// "hidden" is in a comment.
	assert.Empty(t, visited, "Should not visit strings inside comments even if preceded by slash")
}

func TestWalkJSONStrings_Bug_SlashBeforeBlockComment(t *testing.T) {
	// Input contains a division operator (slash) followed by a block comment containing a string.
	input := []byte(`{ "val": 1 / 1 /* "hidden" */ }`)

	visited := []string{}
	WalkJSONStrings(input, func(raw []byte) ([]byte, bool) {
		visited = append(visited, string(raw))
		return nil, false
	})

	assert.Empty(t, visited, "Should not visit strings inside block comments even if preceded by slash")
}
