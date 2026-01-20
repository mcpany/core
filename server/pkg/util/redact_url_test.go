// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No sensitive params",
			input:    "https://example.com/api?query=hello&lang=en",
			expected: "https://example.com/api?query=hello&lang=en",
		},
		{
			name:     "Sensitive param: api_key",
			input:    "https://example.com/api?query=hello&api_key=secret123",
			expected: "https://example.com/api?api_key=%5BREDACTED%5D&query=hello",
		},
		{
			name:     "Sensitive param: token",
			input:    "https://example.com/api?token=abc-def&other=1",
			expected: "https://example.com/api?other=1&token=%5BREDACTED%5D",
		},
		{
			name:     "Sensitive param: password and auth credentials",
			input:    "https://user:pass@example.com/login?password=secret",
			expected: "https://user:%5BREDACTED%5D@example.com/login?password=%5BREDACTED%5D",
		},
		{
			name:     "Multiple sensitive params",
			input:    "https://example.com?api_key=123&secret=456",
			expected: "https://example.com?api_key=%5BREDACTED%5D&secret=%5BREDACTED%5D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.input)
			assert.NoError(t, err)
			result := RedactURL(u)
			// Decode result to check match because param order might change
			resU, _ := url.Parse(result)
			expU, _ := url.Parse(tt.expected)
			assert.Equal(t, expU.Query(), resU.Query())
			assert.Equal(t, expU.Path, resU.Path)
			assert.Equal(t, expU.Host, resU.Host)

			// Check User info
			if expU.User != nil {
				assert.Equal(t, expU.User.String(), resU.User.String())
			}

			if strings.Contains(tt.expected, "REDACTED") {
				assert.Contains(t, result, "%5BREDACTED%5D")
			} else {
				assert.NotContains(t, result, "REDACTED")
			}
		})
	}
}

func TestRedactURLString(t *testing.T) {
	input := "https://example.com?api_key=secret"
	expected := "https://example.com?api_key=%5BREDACTED%5D"
	assert.Contains(t, RedactURLString(input), expected)

	invalid := "::not-a-url"
	assert.Equal(t, invalid, RedactURLString(invalid))
}
