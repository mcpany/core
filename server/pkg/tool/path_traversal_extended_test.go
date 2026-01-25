// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckForPathTraversal_Extended(t *testing.T) {
	tests := []struct {
		input    string
		hasError bool
		desc     string
	}{
		// Basic cases (already covered but good for regression)
		{"safe", false, "safe path"},
		{"..", true, "double dot"},
		{"../", true, "double dot slash"},

		// Encoded cases
		{"%2e%2e", true, "encoded double dot"},
		{"%2e%2e/", true, "encoded double dot slash"},

		// Mixed encoding cases (New vulnerability fix)
		{"%2e.", true, "mixed encoded double dot (first encoded)"},
		{".%2e", true, "mixed encoded double dot (second encoded)"},
		{"%2e./", true, "mixed encoded double dot slash (first encoded)"},
		{".%2e/", true, "mixed encoded double dot slash (second encoded)"},

		// Case insensitivity
		{"%2E%2E", true, "uppercase encoded double dot"},
		{"%2E.", true, "uppercase mixed encoded double dot"},
		{".%2E", true, "uppercase mixed encoded double dot"},

		// False positives check
		{".", false, "single dot"},
		{"...", false, "triple dot (safe usually)"},
		{"%25", false, "percent sign"},
		{"%2e", false, "single encoded dot"},
		{"%2e%2e%2e", true, "encoded triple dot (aggressive check catches this)"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			err := checkForPathTraversal(tt.input)
			if tt.hasError {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
				assert.Contains(t, err.Error(), "path traversal attempt detected")
			} else {
				assert.NoError(t, err, "Expected no error for input: %s", tt.input)
			}
		})
	}
}
