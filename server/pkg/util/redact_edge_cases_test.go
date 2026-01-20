// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

func TestIsSensitiveKey_BugReproduction(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"api_key", true},
		{"private_key", true},
		{"x-api-key", true},
		{"proxy-authorization", true},
		{"token", true}, // Should pass
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := IsSensitiveKey(tt.key); got != tt.expected {
				t.Errorf("IsSensitiveKey(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestRedactJSON_NumericValue(t *testing.T) {
	input := `{"token_id": 12345}`
	expected := `{"token_id": "[REDACTED]"}`

	got := RedactJSON([]byte(input))
	// Remove whitespace for comparison
	if string(got) != expected {
		t.Errorf("RedactJSON(%s) = %s, want %s", input, string(got), expected)
	}
}

func TestRedactJSON_MalformedString(t *testing.T) {
	// Test infinite loop or crash potential in skipString
	input := `{"key": "val` // unclosed string
	got := RedactJSON([]byte(input))
	// Should not hang. Should return input as is or partial?
	// redactJSONFast returns input if malformed?
	// The code breaks loop on malformed.
	if string(got) != input {
		t.Errorf("RedactJSON unclosed string mismatch. Got %q", string(got))
	}
}

func TestRedactJSON_EscapeHandling(t *testing.T) {
	input := `{"token": "\"safe\""}` // Value is "\"safe\"" (string containing quotes)
	// Key is "token". Value should be redacted.
	expected := `{"token": "[REDACTED]"}`
	got := RedactJSON([]byte(input))
	if string(got) != expected {
		t.Errorf("RedactJSON escape handling failed. Got %s, want %s", string(got), expected)
	}
}
