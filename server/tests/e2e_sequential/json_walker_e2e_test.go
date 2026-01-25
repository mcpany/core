// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package e2e_sequential

import (
	"bytes"
	"testing"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
)

// TestWalkJSONStrings_E2E simulates a large document processing scenario
// to verify the robustness of the JSON walker in an "end-to-end" integration context.
// This ensures that the low-level fix works correctly when integrated into the larger build.
func TestWalkJSONStrings_E2E(t *testing.T) {
	// Simulate a large JSON-like log stream or config that might contain stray slashes
	// mixed with comments.

	var input bytes.Buffer
	input.WriteString(`[`)

	// Add some valid items with string values to ensure walker is working
	input.WriteString(`{"id": "valid1"},`)

	// Add the problematic sequence (stray slash followed by comment with string)
	// This represents a corrupted log line or a specific separator usage.
	input.WriteString(` / /* "secret" */ `)

	// Add more items
	input.WriteString(`,{"id": "valid2"}`)
	input.WriteString(`]`)

	data := input.Bytes()

	visited := []string{}
	output := util.WalkJSONStrings(data, func(raw []byte) ([]byte, bool) {
		visited = append(visited, string(raw))
		// Redact anything that looks like "secret"
		if string(raw) == `"secret"` {
			return []byte(`"REDACTED"`), true
		}
		return nil, false
	})

	// Verify valid items are visited
	assert.Contains(t, visited, `"valid1"`)
	assert.Contains(t, visited, `"valid2"`)

	// "secret" is inside a comment, so it should NOT be visited.
	assert.NotContains(t, visited, `"secret"`, "Should not visit string inside comment")

	// Output should not contain REDACTED
	if bytes.Contains(output, []byte(`"REDACTED"`)) {
		t.Errorf("Output contained redaction which means comment was breached: %s", string(output))
	}
}
