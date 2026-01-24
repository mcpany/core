// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"testing"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestRedactSecrets(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		secrets  []string
		expected string
	}{
		{
			name:     "no secrets",
			text:     "hello world",
			secrets:  nil,
			expected: "hello world",
		},
		{
			name:     "empty text",
			text:     "",
			secrets:  []string{"secret"},
			expected: "",
		},
		{
			name:     "single secret",
			text:     "this is a secret message",
			secrets:  []string{"secret"},
			expected: "this is a [REDACTED] message",
		},
		{
			name:     "multiple secrets",
			text:     "user: admin, pass: 12345",
			secrets:  []string{"admin", "12345"},
			expected: "user: [REDACTED], pass: [REDACTED]",
		},
		{
			name:     "overlapping secrets (substrings)",
			text:     "password is SuperSecretPassword",
			secrets:  []string{"SuperSecret", "SuperSecretPassword"},
			expected: "password is [REDACTED]", // Should match longest first
		},
		{
			name:     "empty secret in list",
			text:     "hello world",
			secrets:  []string{"hello", ""},
			expected: "[REDACTED] world", // Should ignore empty secret
		},
		{
			name:     "duplicate secrets",
			text:     "token: abc",
			secrets:  []string{"abc", "abc"},
			expected: "token: [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.RedactSecrets(tt.text, tt.secrets)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactURL(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		expected string
	}{
		{
			name:     "no query params",
			rawURL:   "https://example.com/api/v1/resource",
			expected: "https://example.com/api/v1/resource",
		},
		{
			name:     "non-sensitive query params",
			rawURL:   "https://example.com/search?q=hello&page=1",
			expected: "https://example.com/search?page=1&q=hello", // sorted
		},
		{
			name:     "sensitive query param: api_key",
			rawURL:   "https://example.com/resource?api_key=secret123",
			expected: "https://example.com/resource?api_key=%5BREDACTED%5D",
		},
		{
			name:     "sensitive query param: token",
			rawURL:   "https://example.com/resource?token=abcd&type=user",
			expected: "https://example.com/resource?token=%5BREDACTED%5D&type=user", // sorted
		},
		{
			name:     "sensitive query param: password",
			rawURL:   "https://example.com/login?username=user&password=password123",
			expected: "https://example.com/login?password=%5BREDACTED%5D&username=user", // sorted
		},
		{
			name:     "mixed sensitive and non-sensitive",
			rawURL:   "https://example.com/api?client_id=123&client_secret=secret&response_type=code",
			expected: "https://example.com/api?client_id=123&client_secret=%5BREDACTED%5D&response_type=code", // client_secret is redacted
		},
		{
			name:     "sensitive key in mixed case",
			rawURL:   "https://example.com/api?API_KEY=secret",
			expected: "https://example.com/api?API_KEY=%5BREDACTED%5D",
		},
		{
			name:     "invalid url",
			rawURL:   "://invalid-url",
			expected: "://invalid-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.RedactURL(tt.rawURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}
