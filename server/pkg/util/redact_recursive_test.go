// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

func TestRedactSecrets_Recursive(t *testing.T) {
	// This test reproduces a bug where secrets that are substrings of the
	// redacted placeholder ("[REDACTED]") caused recursive corruption.
	// E.g. "REDACTED", "E", "ACT", etc.

	tests := []struct {
		name     string
		text     string
		secrets  []string
		expected string
	}{
		{
			name:     "Secret is substring of placeholder",
			text:     "This is a secret message.",
			secrets:  []string{"secret", "REDACTED", "E"},
			expected: "This is a [REDACTED] message.",
		},
		{
			name:     "Secret matches placeholder exactly",
			text:     "Start [REDACTED] End",
			secrets:  []string{"[REDACTED]"},
			expected: "Start [REDACTED] End",
		},
		{
			name:     "Multiple secrets overlapping placeholder chars",
			text:     "AB",
			secrets:  []string{"A", "B"},
			expected: "[REDACTED]",
		},
		{
			name:     "Nested secret in placeholder text",
			text:     "The code is REDACTED.",
			secrets:  []string{"REDACTED", "E", "D"},
			expected: "The code is [REDACTED].",
		},
		{
			name:     "Overlapping secrets in text",
			text:     "password123",
			secrets:  []string{"password", "word"},
			expected: "[REDACTED]123",
		},
		{
			name:     "Adjacent secrets",
			text:     "secret1secret2",
			secrets:  []string{"secret1", "secret2"},
			expected: "[REDACTED]", // Merged
		},
		{
			name:     "Adjacent secrets with space",
			text:     "secret1 secret2",
			secrets:  []string{"secret1", "secret2"},
			expected: "[REDACTED] [REDACTED]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := RedactSecrets(tc.text, tc.secrets)
			if got != tc.expected {
				t.Errorf("RedactSecrets(%q, secrets) = %q; want %q", tc.text, got, tc.expected)
			}
		})
	}
}
