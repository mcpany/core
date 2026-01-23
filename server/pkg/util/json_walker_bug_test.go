// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

func TestWalkJSONStrings_Bug_CommentWithPrecedingSlash(t *testing.T) {
	// Logic: The parser sees the first slash in "1 / 1", determines it's not a comment,
	// and then stops looking for comments. It then finds "b" inside the comment
	// and treats it as a string to process.
	input := `{"a": 1 / 1, /* "b": "target" */ "c": "d"}`

	visitor := func(raw []byte) ([]byte, bool) {
		if string(raw) == `"target"` {
			return []byte(`"HIT"`), true
		}
		return nil, false
	}

	expected := `{"a": 1 / 1, /* "b": "target" */ "c": "d"}`
	result := WalkJSONStrings([]byte(input), visitor)

	if string(result) != expected {
		t.Errorf("Test failed! Comment was modified.\nExpected: %s\nGot:      %s", expected, string(result))
	}
}
