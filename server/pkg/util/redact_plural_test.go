// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSensitiveKey_PluralLogic(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"cookies (plural of cookie)", "cookies", true},
		{"tokens (explicitly in list)", "tokens", true},
		{"passwords (explicitly in list)", "passwords", true},

		{"tokenservice (continuation after s)", "tokenservice", false},
		{"cookieservice (continuation after s)", "cookieservice", false},

		{"tokensAuth (CamelCase after s)", "tokensAuth", true},
		{"cookiesAuth (CamelCase after s)", "cookiesAuth", true},

		{"token_service (boundary after key)", "token_service", true},
		{"tokens_service (boundary after s)", "tokens_service", true},

		{"mycookies (suffix match allowed)", "mycookies", true},

		{"not_sensitive", "not_sensitive", false},
		{"not_cookies", "not_cookies", true}, // contains cookie
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSensitiveKey(tt.input)
			assert.Equal(t, tt.want, got, "IsSensitiveKey(%q)", tt.input)
		})
	}
}

func TestRedactJSON_Plurals(t *testing.T) {
	input := `{"cookies": "yummy", "tokens": "secret", "tokenservice": "safe"}`
	want := `{"cookies": "[REDACTED]", "tokens": "[REDACTED]", "tokenservice": "safe"}`

	got := RedactJSON([]byte(input))
	assert.JSONEq(t, want, string(got))
}
