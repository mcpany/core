// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"
)

func TestRedactFast_BufferBoundary(t *testing.T) {
	// Force the use of scanEscapedKeyForSensitive by lowering the limit.
	oldLimit := maxUnescapeLimit
	maxUnescapeLimit = 100
	defer func() { maxUnescapeLimit = oldLimit }()

	// bufSize in scanEscapedKeyForSensitive is 4097.
	// We want to test around the 4096 byte boundary (bufSize - 1).
	const bufBoundary = 4096

	// Sanity check: Ensure "password" is detected in a small buffer
	t.Run("Sanity_Small", func(t *testing.T) {
		key := []byte(`\npassword`)
		if !isKeySensitive(key) {
			t.Errorf("Sanity check failed: password not detected")
		}
	})

	tests := []struct {
		name        string
		prefixLen   int
		sensitive   string
		shouldMatch bool
	}{
		{
			name:        "Match_Before_Boundary",
			prefixLen:   bufBoundary - 10, // "password" fits entirely in first chunk
			sensitive:   "password",
			shouldMatch: true,
		},
		{
			name:        "Match_Crossing_Boundary",
			prefixLen:   bufBoundary - 4, // "pass" | "word" split
			sensitive:   "password",
			shouldMatch: true,
		},
		{
			name:        "Match_After_Boundary_With_Overlap",
			prefixLen:   bufBoundary, // "password" is entirely in second chunk (but pulled into overlap)
			sensitive:   "password",
			shouldMatch: true,
		},
		{
			name:        "No_Match_Crossing_Boundary",
			prefixLen:   bufBoundary - 4,
			sensitive:   "nothing",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			sb.WriteString(`\n`) // 2 chars input -> 1 char unescaped

			// Fill with '.'s (not a start char for sensitive keys)
			// This avoids the loop in scanForSensitiveKeys triggering on every char if we used 'a'.
			padding := tt.prefixLen - 1
			if padding > 0 {
				sb.WriteString(strings.Repeat(".", padding))
			}

			// Append the sensitive word
			sb.WriteString(tt.sensitive)

			// Append some trailing content to ensure we flush the buffer
			// Use '!' to avoid boundary check failures (unlike 'z' which looks like word continuation)
			sb.WriteString(strings.Repeat("!", 100))

			keyContent := []byte(sb.String())

			// Call the internal function
			got := isKeySensitive(keyContent)

			if got != tt.shouldMatch {
				t.Errorf("isKeySensitive() = %v, want %v (prefixLen=%d, sensitive=%q)", got, tt.shouldMatch, tt.prefixLen, tt.sensitive)
			}
		})
	}
}
