// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"testing"
)

func TestWalkJSONStrings_BugRepro(t *testing.T) {
	// Case: A slash that is NOT a comment, followed by a comment containing a string.
	// The parser should NOT visit the string inside the comment.
	// We simulate a scenario where '/' is used as a separator or trailing character.
	// The stray slash and the comment must appear between two strings (or start/end).
	// We use an array to avoid any potential confusion with keys (though keys are skipped anyway).
	input := []byte(`[1] / /* "secret" */`)

	visited := []string{}
	output := WalkJSONStrings(input, func(raw []byte) ([]byte, bool) {
		visited = append(visited, string(raw))
		return []byte(`"REDACTED"`), true
	})

	for _, v := range visited {
		if v == `"secret"` {
			t.Errorf("Bug reproduced: visited string inside comment: %s", v)
		}
	}

	if bytes.Contains(output, []byte(`"REDACTED"`)) {
		t.Errorf("Bug reproduced: input was modified inside comment: %s", string(output))
	}
}
