// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		contains []string // used when order might differ
	}{
		{
			name:     "Empty URL",
			input:    "",
			expected: "",
		},
		{
			name:     "No query params",
			input:    "http://example.com/path",
			expected: "http://example.com/path",
		},
		{
			name:     "Safe query params",
			input:    "http://example.com/path?foo=bar&baz=qux",
			expected: "http://example.com/path?foo=bar&baz=qux",
		},
		{
			name:     "Sensitive query param (api_key)",
			input:    "http://example.com/path?api_key=secret123&foo=bar",
			expected: "http://example.com/path?api_key=[REDACTED]&foo=bar",
		},
		{
			name:     "Sensitive query param (token)",
			input:    "http://example.com/path?token=secret123",
			expected: "http://example.com/path?token=[REDACTED]",
		},
		{
			name:     "Mixed sensitive and safe params",
			input:    "http://example.com/api?user=bob&password=secret&lang=en",
			contains: []string{"password=[REDACTED]", "user=bob", "lang=en"},
		},
		{
			name:     "URL with user info (password)",
			input:    "http://user:password@example.com/path",
			expected: "http://user:[REDACTED]@example.com/path",
		},
		{
			name:     "URL with user info and sensitive query",
			input:    "http://user:password@example.com/path?secret=foo",
			expected: "http://user:[REDACTED]@example.com/path?secret=[REDACTED]",
		},
		{
			name:     "Invalid URL (fallback to RedactDSN behavior)",
			input:    "postgres://user:password@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactURL(tt.input)
			if len(tt.contains) > 0 {
				for _, c := range tt.contains {
					assert.Contains(t, result, c)
				}
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
