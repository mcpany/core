// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"
)

// TestRedactFast_BufferBoundary tests the buffer handling logic in scanEscapedKeyForSensitive.
// The internal buffer size is 4097, flushing at 4096 bytes.
// We want to ensure that sensitive keys split across this boundary are correctly detected
// due to the overlap buffer mechanism.
func TestRedactFast_BufferBoundary(t *testing.T) {
	// Save original limit and restore after test
	oldLimit := maxUnescapeLimit
	defer func() { maxUnescapeLimit = oldLimit }()

	// Set limit low to force the streaming path (scanEscapedKeyForSensitive)
	// instead of json.Unmarshal or stack buffer.
	maxUnescapeLimit = 100

	const bufSize = 4097
	const flushSize = bufSize - 1 // 4096

	tests := []struct {
		name        string
		prefixLen   int
		keyword     string
		shouldMatch bool
	}{
		{
			name:        "Split_Keyword_Across_Boundary",
			prefixLen:   flushSize - 4, // 4092 bytes of prefix. "pass" (4 bytes) fills 4092-4096. "word" starts at 4096.
			keyword:     "password",
			shouldMatch: true,
		},
		{
			name:        "Keyword_Ends_At_Boundary",
			prefixLen:   flushSize - 8, // 4088 bytes. "password" (8 bytes) fills 4088-4096.
			keyword:     "password",
			shouldMatch: true,
		},
		{
			name:        "Keyword_Starts_At_Boundary",
			prefixLen:   flushSize, // 4096 bytes. "password" starts at 4096 (start of next chunk).
			keyword:     "password",
			shouldMatch: true,
		},
		{
			name:        "Keyword_In_Overlap_Zone",
			prefixLen:   flushSize - 20, // 4076 bytes. "password" at 4076-4084. Inside last 64 bytes (overlap).
			keyword:     "password",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Construct the key content.
			// We need to trigger unescaping, so we include a backslash sequence.
			// We use \n at the start, which unescapes to 1 byte ('\n').
			var sb strings.Builder
			sb.WriteString(`\n`) // 1 byte unescaped

			// Fill the rest of the prefix with 'a's.
			// Total prefix length should be tt.prefixLen.
			// We already have 1 byte from \n.
			padLen := tt.prefixLen - 1
			for i := 0; i < padLen; i++ {
				sb.WriteByte('a')
			}

			// Append the keyword
			sb.WriteString(tt.keyword)

			keyContent := []byte(sb.String())

			// Verify that our construction logic is correct regarding length
			// Note: isKeySensitive takes the RAW (escaped) key content.
			// But scanEscapedKeyForSensitive processes the UNESCAPED content.
			// Our construction makes unescaped length = 1 + padLen + len(keyword).
			// 1 + (tt.prefixLen - 1) + len(keyword) = tt.prefixLen + len(keyword).
			// So "password" starts at index tt.prefixLen in the unescaped stream.

			// Ensure we are triggering the right path
			if len(keyContent) <= maxUnescapeLimit {
				t.Fatalf("Test setup error: key length %d is not > maxUnescapeLimit %d", len(keyContent), maxUnescapeLimit)
			}

			// Check if sensitive
			got := isKeySensitive(keyContent)
			if got != tt.shouldMatch {
				t.Errorf("isKeySensitive() = %v, want %v (prefixLen=%d)", got, tt.shouldMatch, tt.prefixLen)
			}
		})
	}
}
