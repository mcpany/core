// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOAuthConfig(t *testing.T) {
	testCases := []struct {
		name        string
		config      *OAuth2Config
		expectError bool
		expected    *OAuth2Config
	}{
		{
			name:        "Empty config",
			config:      &OAuth2Config{},
			expectError: false,
			expected:    &OAuth2Config{},
		},
		{
			name: "Valid config",
			config: &OAuth2Config{
				IssuerURL: "test-issuer",
				Audience:  "test-audience",
			},
			expectError: false,
			expected: &OAuth2Config{
				IssuerURL: "test-issuer",
				Audience:  "test-audience",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.config)
		})
	}
}
