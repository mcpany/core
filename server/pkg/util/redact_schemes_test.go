// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactDSN_Schemes(t *testing.T) {
	// These schemes should NOT be redacted when they look like host:service (no @)
	// because they are transport protocols often used with service names as ports.

	testCases := []struct {
		input    string
		expected string
	}{
		{"ws://localhost:web", "ws://localhost:web"},
		{"wss://localhost:web", "wss://localhost:web"},
		{"grpc://localhost:web", "grpc://localhost:web"},
		// Keep existing behavior for http/s
		{"http://localhost:web", "http://localhost:web"},
		{"https://localhost:web", "https://localhost:web"},
		// But redact if @ is present (implies user:pass)
		{"ws://user:pass@localhost:web", "ws://user:[REDACTED]@localhost:web"},
		{"grpc://user:pass@localhost:web", "grpc://user:[REDACTED]@localhost:web"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, RedactDSN(tc.input), "Failed for input: %s", tc.input)
	}
}
