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
		{
			name:     "adjacent secrets",
			text:     "secret1secret2",
			secrets:  []string{"secret1", "secret2"},
			expected: "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.RedactSecrets(tt.text, tt.secrets)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactDSN(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		expected string
	}{
		{
			name:     "postgres standard",
			dsn:      "postgres://user:password@localhost:5432/dbname",
			expected: "postgres://user:[REDACTED]@localhost:5432/dbname",
		},
		{
			name:     "mysql standard",
			dsn:      "mysql://user:password@tcp(localhost:3306)/dbname",
			expected: "mysql://user:[REDACTED]@tcp(localhost:3306)/dbname",
		},
		{
			name:     "redis no user",
			dsn:      "redis://:password@localhost:6379/0",
			expected: "redis://:[REDACTED]@localhost:6379/0",
		},
		{
			name:     "password with special chars",
			dsn:      "postgres://user:p@ss:w0rd@localhost:5432/dbname",
			expected: "postgres://user:[REDACTED]@localhost:5432/dbname",
		},
		{
			name:     "invalid port leak (go error)",
			dsn:      `parse "postgres://user:password@localhost:5432/dbname": invalid port ":password"`,
			expected: `parse "postgres://user:[REDACTED]@localhost:5432/dbname": invalid port ":[REDACTED]"`,
		},
		{
			name:     "http url (no auth)",
			dsn:      "http://example.com/path",
			expected: "http://example.com/path",
		},
		{
			name:     "http url with auth",
			dsn:      "http://user:password@example.com/path",
			expected: "http://user:[REDACTED]@example.com/path",
		},
		{
			name:     "mailto (ignored)",
			dsn:      "mailto:bob@example.com",
			expected: "mailto:bob@example.com",
		},
		{
			name:     "complex password with @",
			dsn:      "postgres://user:p@ssword@localhost:5432/dbname",
			expected: "postgres://user:[REDACTED]@localhost:5432/dbname",
		},
		{
			name:     "password with colon",
			dsn:      "postgres://user:pass:word@localhost:5432/dbname",
			expected: "postgres://user:[REDACTED]@localhost:5432/dbname",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.RedactDSN(tt.dsn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "flat map no secrets",
			input:    map[string]interface{}{"name": "Alice", "age": 30},
			expected: map[string]interface{}{"name": "Alice", "age": 30},
		},
		{
			name:     "flat map with secret",
			input:    map[string]interface{}{"name": "Alice", "password": "supersecret"},
			expected: map[string]interface{}{"name": "Alice", "password": "[REDACTED]"},
		},
		{
			name:     "nested map with secret",
			input:    map[string]interface{}{"user": map[string]interface{}{"name": "Alice", "token": "abcdef"}},
			expected: map[string]interface{}{"user": map[string]interface{}{"name": "Alice", "token": "[REDACTED]"}},
		},
		{
			name:     "slice inside map with secrets",
			input: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{"id": 1, "secret": "one"},
					map[string]interface{}{"id": 2, "secret": "two"},
				},
			},
			expected: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{"id": 1, "secret": "[REDACTED]"},
					map[string]interface{}{"id": 2, "secret": "[REDACTED]"},
				},
			},
		},
		{
			name:     "mixed nested structure",
			input:    map[string]interface{}{"meta": map[string]interface{}{"key": "val", "api_key": "123"}},
			expected: map[string]interface{}{"meta": map[string]interface{}{"key": "val", "api_key": "[REDACTED]"}},
		},
		{
			name:     "non-sensitive keys preserved",
			input:    map[string]interface{}{"public_key": "pub123", "description": "safe"},
			expected: map[string]interface{}{"public_key": "pub123", "description": "safe"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.RedactMap(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		// Lowercase matches
		{"api_key", "api_key", true},
		{"token", "token", true},
		{"secret", "secret", true},
		{"password", "password", true},
		{"passwd", "passwd", true},
		{"credential", "credential", true},
		{"auth", "auth", true},
		{"private_key", "private_key", true},
		{"authorization", "authorization", true},
		{"cookie", "cookie", true},
		{"set-cookie", "set-cookie", true},
		{"x-api-key", "x-api-key", true},

		// Plural forms
		{"passwords", "passwords", true},
		{"tokens", "tokens", true},
		{"api_keys", "api_keys", true},
		{"credentials", "credentials", true},
		{"secrets", "secrets", true},

		// Mixed Case / CamelCase
		{"ApiKey", "ApiKey", true},
		{"APIKey", "APIKey", true},
		{"Token", "Token", true},
		{"Secret", "Secret", true},
		{"Password", "Password", true},
		{"authToken", "authToken", true},
		{"AuthToken", "AuthToken", true},

		// Snake Case with variations
		{"my_api_key", "my_api_key", true},
		{"user_password", "user_password", true},
		{"app_secret", "app_secret", true},

		// Boundary conditions
		{"authentication", "authentication", true},
		{"author", "author", false}, // "auth" is prefix, but "or" continues
		{"authority", "authority", false},
		{"authentic", "authentic", false},

		// Negative cases
		{"user", "user", false},
		{"name", "name", false},
		{"id", "id", false},
		{"description", "description", false},
		{"public_key", "public_key", false},
		{"my_token_id", "my_token_id", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.IsSensitiveKey(tt.key)
			assert.Equal(t, tt.expected, result, "key: %s", tt.key)
		})
	}
}

func TestRedactJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "flat object no secrets",
			input:    `{"name": "Alice", "age": 30}`,
			expected: `{"name": "Alice", "age": 30}`,
		},
		{
			name:     "flat object with secret",
			input:    `{"name": "Alice", "password": "supersecret"}`,
			expected: `{"name": "Alice", "password": "[REDACTED]"}`,
		},
		{
			name:     "nested object with secret",
			input:    `{"user": {"name": "Alice", "token": "abcdef"}}`,
			expected: `{"user": {"name": "Alice", "token": "[REDACTED]"}}`,
		},
		{
			name:     "array with secrets",
			input:    `[{"id": 1, "secret": "one"}, {"id": 2, "secret": "two"}]`,
			expected: `[{"id": 1, "secret": "[REDACTED]"}, {"id": 2, "secret": "[REDACTED]"}]`,
		},
		{
			name:     "mixed structure",
			input:    `{"data": [{"key": "value"}, {"api_key": "12345"}]}`,
			expected: `{"data": [{"key": "value"}, {"api_key": "[REDACTED]"}]}`,
		},
		{
			name:     "malformed JSON",
			input:    `{this is not json}`,
			expected: `{this is not json}`,
		},
		// Whitespace and comments (supported by redactJSONFast)
        {
            name: "whitespace and comments",
            input: `
            // comment before
            {
                "password": "secret", // comment after
                /* block comment */
                "token": "123"
            }`,
            expected: `
            // comment before
            {
                "password": "[REDACTED]", // comment after
                /* block comment */
                "token": "[REDACTED]"
            }`,
        },
        // Escaped quotes in keys and values
        {
            name: "escaped quotes",
            input: `{"key\"with\"quotes": "value", "secret": "val\"ue"}`,
            expected: `{"key\"with\"quotes": "value", "secret": "[REDACTED]"}`,
        },
        // Unicode escapes
        {
            name: "unicode escapes",
            input: `{"\u0070\u0061\u0073\u0073\u0077\u006f\u0072\u0064": "secret"}`, // "password"
            expected: `{"\u0070\u0061\u0073\u0073\u0077\u006f\u0072\u0064": "[REDACTED]"}`,
        },
        // Numbers and other types
        {
            name: "numbers and bools",
            input: `{"secret": 12345, "token": true, "key": null}`,
            expected: `{"secret": "[REDACTED]", "token": "[REDACTED]", "key": null}`,
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.RedactJSON([]byte(tt.input))
            // Normalize spaces for comparison if needed, but RedactJSON should preserve exact bytes for non-redacted parts.
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestRedactURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "no params",
			url:      "http://example.com/path",
			expected: "http://example.com/path",
		},
		{
			name:     "safe params",
			url:      "http://example.com/path?id=123&name=alice",
			expected: "http://example.com/path?id=123&name=alice",
		},
		{
			name:     "sensitive params",
			url:      "http://example.com/path?api_key=secret123&token=abc",
			expected: "http://example.com/path?api_key=[REDACTED]&token=[REDACTED]",
		},
		{
			name:     "mixed params",
			url:      "http://example.com/path?id=123&api_key=secret",
			expected: "http://example.com/path?api_key=[REDACTED]&id=123", // Query param order might change, but that's expected behavior of url.Values.Encode()
		},
		{
			name:     "invalid url",
			url:      "://invalid",
			expected: "://invalid",
		},
		{
			name:     "sensitive param case insensitive",
			url:      "http://example.com?APIKey=secret",
			expected: "http://example.com?APIKey=[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.RedactURL(tt.url)
			if tt.name == "mixed params" {
				// Order of query params is not guaranteed
				assert.Contains(t, result, "api_key=[REDACTED]")
				assert.Contains(t, result, "id=123")
				assert.Contains(t, result, "http://example.com/path?")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
