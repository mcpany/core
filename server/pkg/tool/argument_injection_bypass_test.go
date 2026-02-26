// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestValidateSafePathAndInjection_Bypass(t *testing.T) {
	// Restore IsSafeURL after test
	originalIsSafeURL := validation.IsSafeURL
	defer func() { validation.IsSafeURL = originalIsSafeURL }()

	// Set strict IsSafeURL that mimics real behavior for testing SSRF
	validation.IsSafeURL = func(urlStr string) error {
		if strings.Contains(urlStr, "localhost") || strings.Contains(urlStr, "127.0.0.1") {
			return fmt.Errorf("unsafe URL: localhost/loopback not allowed")
		}
		return nil
	}

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
			name:     "Path Traversal with Leading Space",
			input:    " ../etc/passwd",
			isDocker: false,
		},
		{
			name:     "Local File Access with Leading Space",
			input:    " /etc/passwd", // Absolute path check
			isDocker: false,
		},
		{
			name:     "Double Encoded Hyphen (Argument Injection)",
			input:    "%252d",
			isDocker: false,
		},
		{
			name:     "Triple Encoded Hyphen (Argument Injection)",
			input:    "%25252d",
			isDocker: false,
		},
		{
			name:     "Encoded URL Scheme (SSRF)",
			input:    "http%3a%2f%2flocalhost",
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
