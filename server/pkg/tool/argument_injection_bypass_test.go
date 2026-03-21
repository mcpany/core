// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSafePathAndInjection_Bypass(t *testing.T) {
	// These inputs should be BLOCKED.
	tests := []struct {
		name     string
		input    string
		isDocker bool
	}{
		{
			name:     "Argument Injection with Leading Space",
			input:    " -dangerous",
			isDocker: false,
		},
		{
			name:     "Argument Injection with URL Encoded Leading Space",
			input:    "%20-dangerous",
			isDocker: false,
		},
		{
			name:     "Argument Injection with Leading Tab",
			input:    "\t-dangerous",
			isDocker: false,
		},
		{
			name:     "Argument Injection with Plus and Leading Space",
			input:    " +dangerous",
			isDocker: false,
		},
		{
			name:     "Path Traversal with Leading Space",
			input:    " ../etc/passwd",
			isDocker: false,
		},
		{
			name:     "Path Traversal with URL Encoded Leading Space",
			input:    "%20../etc/passwd",
			isDocker: false,
		},
		{
			name:     "Path Traversal with Leading Tab",
			input:    "\t../etc/passwd",
			isDocker: false,
		},
		{
			name:     "Path Traversal with Trailing Space",
			input:    "../etc/passwd ",
			isDocker: false,
		},
		{
			name:     "Local File Access with Leading Space",
			input:    " /etc/passwd", // Absolute path check
			isDocker: false,
		},
		{
			name:     "Local File Access with URL Encoded Leading Space",
			input:    "%20/etc/passwd", // Absolute path check
			isDocker: false,
		},
		{
			name:     "Local File Access with Trailing Space",
			input:    "/etc/passwd ", // Absolute path check
			isDocker: false,
		},
		{
			name:     "Local File Access with file:// Scheme and Leading Space",
			input:    " file:///etc/passwd", // file scheme check
			isDocker: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSafePathAndInjection(tt.input, tt.isDocker, "generic-tool")
			assert.Error(t, err, "Should block input: %q", tt.input)
		})
	}
}
