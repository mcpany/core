// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckForPathTraversal_MixedEncoding(t *testing.T) {
	// These patterns decode to ".." but might be missed by simple "%2e%2e" checks
	testCases := []struct {
		input    string
		shouldError bool
	}{
		{"%2e.", true},
		{".%2e", true},
		{"%2E.", true},
		{".%2E", true},
		{"foo/%2e.", true},
		{"%2e./bar", true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			err := checkForPathTraversal(tc.input)
			if tc.shouldError {
				assert.Error(t, err, "Input %q should be detected as traversal", tc.input)
			} else {
				assert.NoError(t, err, "Input %q should NOT be detected as traversal", tc.input)
			}
		})
	}
}
