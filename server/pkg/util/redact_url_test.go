// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net/url"
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
			input:    "https://example.com/api?foo=bar&baz=qux",
			expected: "https://example.com/api?foo=bar&baz=qux",
		},
		{
			name:     "Sensitive param api_key",
			input:    "https://example.com/api?api_key=12345&foo=bar",
			expected: "https://example.com/api?api_key=%5BREDACTED%5D&foo=bar",
		},
		{
			name:     "Sensitive param token",
			input:    "https://example.com/api?token=secret&user=bob",
			expected: "https://example.com/api?token=%5BREDACTED%5D&user=bob",
		},
		{
			name:     "Mixed sensitive and non-sensitive",
			input:    "https://example.com/api?id=1&password=pass&debug=true",
			expected: "https://example.com/api?debug=true&id=1&password=%5BREDACTED%5D",
		},
		{
			name:     "URL without query",
			input:    "https://example.com/api",
			expected: "https://example.com/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.input)
			assert.NoError(t, err)

			got := RedactURL(u)

			// Since map iteration order is random, query params might be reordered.
			// url.Values.Encode() sorts keys by key, so expected should be sorted by key.
			// Let's verify by parsing the result back and comparing values.

			gotU, err := url.Parse(got)
			assert.NoError(t, err)

			expectedU, err := url.Parse(tt.expected)
			assert.NoError(t, err)

			assert.Equal(t, expectedU.Scheme, gotU.Scheme)
			assert.Equal(t, expectedU.Host, gotU.Host)
			assert.Equal(t, expectedU.Path, gotU.Path)

			expectedQ := expectedU.Query()
			gotQ := gotU.Query()

			assert.Equal(t, expectedQ, gotQ)
		})
	}
}
