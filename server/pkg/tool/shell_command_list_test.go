// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsShellCommand_ExtendedList(t *testing.T) {
	tests := []struct {
		cmd      string
		expected bool
	}{
		// Existing
		{"python", true},
		{"bash", true},
		{"sh", true},
		{"python3.10", true},

		// New additions
		{"tar", false},
		{"find", false},
		{"xargs", false},
		{"make", false},
		{"npm", false},
		{"npx", false},
		{"bunx", false},
		{"go", false},
		{"cargo", false},
		{"pip", false},
		{"gradle", false},
		{"mvn", false},
		{"ant", false},

		// Safe(r) commands (not in blocklist)
		{"ls", false},
		{"grep", false},
		{"cat", false},
		{"echo", false}, // echo is shell builtin but /bin/echo exists. Is it dangerous? Only if it interprets backslashes or something, but usually fine as long as not `sh -c echo ...`
		{"my-app", false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			assert.Equal(t, tt.expected, isShellCommand(tt.cmd), "isShellCommand(%q) should be %v", tt.cmd, tt.expected)
		})
	}
}
