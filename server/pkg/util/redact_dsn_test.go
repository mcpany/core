// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactDSN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard postgres",
			input:    "postgres://user:password@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
		{
			name:     "password with colon",
			input:    "postgres://user:pass:word@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
		{
			name:     "password with multiple colons",
			input:    "postgres://user:p1:p2:p3@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
		{
			name:     "password with at sign (raw)",
			input:    "postgres://user:pass@word@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
		{
			name:     "password with at sign (encoded)",
			input:    "postgres://user:pass%40word@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
		{
			name:     "no password",
			input:    "mysql://user@host:3306/db",
			expected: "mysql://user@host:3306/db",
		},
		{
			name:     "complex scheme",
			input:    "mongodb+srv://user:password@cluster.mongodb.net/db",
			expected: "mongodb+srv://user:[REDACTED]@cluster.mongodb.net/db",
		},
		{
			name:     "not a url",
			input:    "not-a-url",
			expected: "not-a-url",
		},
		{
			name:     "url without user info",
			input:    "http://example.com/path",
			expected: "http://example.com/path",
		},
		{
			name:     "redis with password only",
			input:    "redis://:password@localhost:6379",
			expected: "redis://:[REDACTED]@localhost:6379",
		},
		{
			name:     "invalid url with control char (fallback)",
			input:    "postgres://user:password@host/db\n",
			expected: "postgres://user:[REDACTED]@host/db\n",
		},
		{
			name:     "mysql generic driver format (regression check)",
			input:    "user:password@tcp(host:3306)/db",
			expected: "user:[REDACTED]@tcp(host:3306)/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactDSN(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
