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
		// New cases for regex fallback improvements
		{
			name:     "fallback: password with colon (invalid scheme)",
			input:    "post gres://user:pass:word@localhost:5432/db",
			expected: "post gres://user:[REDACTED]@localhost:5432/db",
		},
		{
			name:     "fallback: schemeless with colon in password",
			input:    "user:pass:word@host",
			expected: "user:[REDACTED]@host",
		},
		{
			name:     "fallback: password with space",
			input:    "postgres://user:pass word@localhost:5432/db",
			expected: "postgres://user:[REDACTED] word@localhost:5432/db",
		},
		{
			name:     "fallback: empty password",
			input:    "user:@host",
			expected: "user:[REDACTED]@host",
		},
		{
			name:     "fallback: password starting with slash",
			input:    "user:/pass@host",
			expected: "user:[REDACTED]@host",
		},
		{
			name:     "fallback: password starting with double slash (leak/limitation - looks like scheme)",
			input:    "user://pass@host",
			expected: "user://pass@host",
		},
		{
			name:     "fallback: password with colon and at sign",
			input:    "user:pass:word@part@host",
			expected: "user:[REDACTED]@host",
		},
		{
			name:     "fallback: empty user with password",
			input:    "postgres://:password@host",
			expected: "postgres://:[REDACTED]@host",
		},
		{
			name:     "mailto scheme (whitelist)",
			input:    "mailto:bob@example.com",
			expected: "mailto:bob@example.com",
		},
		{
			name:     "mailto scheme case insensitive (whitelist)",
			input:    "MAILTO:bob@example.com",
			expected: "MAILTO:bob@example.com",
		},
		{
			name:     "multiple DSNs in string",
			input:    "Connect to mysql://user1:secret1@host1 and postgres://user2:secret2@host2",
			expected: "Connect to mysql://user1:[REDACTED]@host1 and postgres://user2:[REDACTED]@host2",
		},
		{
			name:     "path/query containing @",
			input:    "postgres://user:password@host:invalidport/db?email=foo@bar.com",
			expected: "postgres://user:[REDACTED]@host:invalidport/db?email=foo@bar.com",
		},
		// New cases for named ports in http/https (Bug fix)
		{
			name:     "http named port",
			input:    "http://myservice:web",
			expected: "http://myservice:web",
		},
		{
			name:     "https named port",
			input:    "https://myservice:web",
			expected: "https://myservice:web",
		},
		{
			name:     "redis empty host password",
			input:    "redis://:secret",
			expected: "redis://:[REDACTED]",
		},
		{
			name:     "scylla host list",
			input:    "scylla://node1:9042,node2:9042",
			expected: "scylla://node1:9042,node2:9042",
		},
		{
			name:     "http credentials with at",
			input:    "http://user:secret@host",
			expected: "http://user:[REDACTED]@host",
		},
		{
			name:     "http credentials without at (treated as host:port)",
			input:    "http://user:secret",
			expected: "http://user:secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactDSN(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
