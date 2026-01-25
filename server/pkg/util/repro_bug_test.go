// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

func TestWalkJSONStrings_BugRepro(t *testing.T) {
	input := `[ 1/2, /* "hidden" */ "visible" ]`
	visitedHidden := false
	visitedVisible := false

	visitor := func(raw []byte) ([]byte, bool) {
		s := string(raw)
		t.Logf("Visited: %s", s)
		if s == `"hidden"` {
			visitedHidden = true
		}
		if s == `"visible"` {
			visitedVisible = true
		}
		return nil, false
	}

	WalkJSONStrings([]byte(input), visitor)

	if visitedHidden {
		t.Errorf("Bug reproduced: Visited string inside comment!")
	}
	if !visitedVisible {
		t.Errorf("Bug side-effect: Failed to visit visible string")
	}
}
