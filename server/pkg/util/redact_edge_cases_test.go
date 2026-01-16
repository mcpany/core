// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import "testing"

func TestRedact_EdgeCases(t *testing.T) {
	// Test boundary conditions in scanForSensitiveKeys -> checkPotentialMatch
	t.Run("WordBoundaries", func(t *testing.T) {
		tests := []struct {
			input string
			want  bool
		}{
			{"author", false},      // auth is prefix, but 'o' continues word
			{"authority", false},
			{"authentication", true}, // 'authentication' is in the sensitive list
			{"AUTH", true},
			{"AUTHOR", false},      // continuation
			{"AuthToken", true},    // CamelCase boundary
			{"AUTH_TOKEN", true},   // underscore boundary
			{"my_auth", true},      // found inside
			{"token", true},
			{"tokens", true},
			{"tokenization", false}, // token is prefix, 'i' continues
		}
		for _, tt := range tests {
			got := scanForSensitiveKeys([]byte(tt.input), false)
			if got != tt.want {
				t.Errorf("scanForSensitiveKeys(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	})

	// Test isKey edge cases
	t.Run("isKey_EdgeCases", func(t *testing.T) {
		// 1. Hit maxScan limit
		// We need an input starting at startOffset, without '"' or '\\', longer than maxScan (256).
		longStr := make([]byte, 300)
		for i := 0; i < 300; i++ {
			longStr[i] = 'a'
		}
		// isKey starts scanning at startOffset.
		// It returns true (conservative) if limit reached.
		if !isKey(longStr, 0) {
			t.Errorf("expected true (conservative) for long string")
		}

		// 2. Escape sequence
		// If it finds '\\', it returns true (conservative).
		escStr := []byte(`abc\def`)
		if !isKey(escStr, 0) {
			t.Errorf("expected true for string with escape")
		}

		// 3. Quote but no colon (EOF)
		noColon := []byte(`abc"   `)
		if isKey(noColon, 0) {
			t.Errorf("expected false for quote with no colon")
		}
	})

	// Test scanJSONForSensitiveKeys
	t.Run("scanJSONForSensitiveKeys", func(t *testing.T) {
		tests := []struct {
			input string
			want  bool
		}{
			{`{"token": "val"}`, true},
			// "nottoken" contains "token", and since we don't check previous character (suffix match allowed),
			// it is considered sensitive.
			{`{"nottoken": "val"}`, true},
			{`{"key": "token"}`, false}, // token in value, not key
			{`"token"`, false},          // just string, no colon
			{`"token":`, true},          // key
		}
		for _, tt := range tests {
			got := scanJSONForSensitiveKeys([]byte(tt.input))
			if got != tt.want {
				t.Errorf("scanJSONForSensitiveKeys(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	})
}
