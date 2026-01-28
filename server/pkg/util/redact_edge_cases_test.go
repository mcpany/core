// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		input string
		wantRedacted bool
	}{
		// Basic
		{"basic", `{"token": "secret"}`, true},

		// Unclosed string - robustly handled
		{"unclosed string", `{"token": "secret`, true},

		// Comment handling
		{"comment before key", `{ // comment
"token": "secret"}`, true},
		{"comment inside value", `{"token": 123 // comment
}`, true},
		{"slash in string", `{"key": "http://example.com", "token": "secret"}`, true},

		// Weird spacing
		{"spaces", ` {  "token"  :   "secret"  } `, true},

		// Empty keys/values
		{"empty key", `{"": "val", "token": "secret"}`, true},
		{"empty value", `{"token": ""}`, true},

		// Nested
		{"nested array", `{"a": [{"token": "secret"}]}`, true},

		// Keys with escapes
		{"escaped quote in key", `{"to\"ken": "secret"}`, false}, // "to\"ken" != "token"
		{"escaped backslash in key", `{"to\\ken": "secret"}`, false},

		// Sensitive key variants
		{"auth", `{"auth": "val"}`, true},
		{"Auth", `{"Auth": "val"}`, true}, // Case insensitive
		{"AUTH", `{"AUTH": "val"}`, true},

		// Boundary check
		{"author", `{"author": "val"}`, false}, // Should not redact
		{"auth_token", `{"auth_token": "val"}`, true}, // Should redact

		// MyAuth - matched because "Auth" is sensitive and recognized as component
		{"MyAuth", `{"MyAuth": "secret"}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactJSON([]byte(tt.input))
			s := string(got)
			if tt.wantRedacted {
				assert.Contains(t, s, "[REDACTED]", "Should be redacted")
				assert.NotContains(t, s, "secret", "Should not contain secret")
			} else {
				assert.NotContains(t, s, "[REDACTED]", "Should not be redacted")
			}
		})
	}
}
