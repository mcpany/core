// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"testing"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestRedactURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple url",
			input:    "http://example.com/api",
			expected: "http://example.com/api",
		},
		{
			name:     "sensitive query param",
			input:    "http://example.com/api?api_key=secret123",
			expected: "http://example.com/api?api_key=%5BREDACTED%5D",
		},
		{
			name:     "multiple query params",
			input:    "http://example.com/api?user=bob&token=xyz",
			expected: "http://example.com/api?token=%5BREDACTED%5D&user=bob", // ordering might change
		},
		{
			name:     "user info in url",
			input:    "http://user:password@example.com",
			expected: "http://user:%5BREDACTED%5D@example.com",
		},
		{
			name:     "fragment with sensitive data",
			input:    "http://example.com/#access_token=123",
			expected: "http://example.com/#access_token=%5BREDACTED%5D",
		},
		{
			name:     "invalid url",
			input:    "://invalid",
			expected: "://invalid",
		},
		{
			name:     "non-sensitive query",
			input:    "http://example.com?search=hello",
			expected: "http://example.com?search=hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.RedactURL(tt.input)
			// For query parameters, order is not guaranteed, so we might need a more robust check if exact match fails
			// But for simple cases with 1 or 2 params, let's see.
			// url.Values.Encode sorts by key.
			assert.Equal(t, tt.expected, result)
		})
	}
}
