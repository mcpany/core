// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWalkJSONStrings_Bug_CommentPrecededBySlash(t *testing.T) {
	// The bug: If a non-comment slash appears before a comment, the comment is not detected,
	// and content inside the comment (like quotes) is processed as JSON string.

	input := []byte(`{
		"a": 10 / 2, // "password": "secret"
		"b": "value"
	}`)

	// We use a visitor that just echoes the string.
	// If "password" is detected as a string, the visitor will be called for it.
	// We can track which strings are visited.

	visited := []string{}
	WalkJSONStrings(input, func(raw []byte) ([]byte, bool) {
		visited = append(visited, string(raw))
		return nil, false
	})

	// Expected: "a", "b", "value".
	// "password" and "secret" should NOT be visited because they are in a comment.
	// Note: WalkJSONStrings visits VALUES.

	// If bug exists, "secret" WILL be in visited because "password" is seen as key (followed by :)
	// and "secret" is seen as value.

	assert.NotContains(t, visited, `"secret"`, "String inside comment should not be visited")
}
