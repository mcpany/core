// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"

)

func TestRedactBugFalsePositiveAtChunkBoundary(t *testing.T) {
	// Save original limit and restore it after test
	originalLimit := maxUnescapeLimit
	defer func() { maxUnescapeLimit = originalLimit }()

	// Set limit small to force scanEscapedKeyForSensitive
	maxUnescapeLimit = 100

	// We want to construct a key that, when unescaped, has "auth" at the end of the first 4096-byte chunk.
	// Buffer size in scanEscapedKeyForSensitive is 4096.
	// We want "auth" to be at indices 4092, 4093, 4094, 4095.
	// The next char 'o' (from "authority") should be at index 4096 (start of next chunk).

	// Unescaped content structure:
	// prefix: 4092 chars.
	// "authority": 9 chars.
	// Total unescaped length: 4101.

	// We use an escape sequence at the start to ensure scanEscapedKeyForSensitive is triggered.
	// \n is 1 byte when unescaped.
	// So we need 4091 'x's.

	sb := strings.Builder{}
	sb.WriteString(`\n`) // Unescapes to 1 byte
	for i := 0; i < 4091; i++ {
		sb.WriteByte('x')
	}
	sb.WriteString("authority")

	key := sb.String()

	// Construct JSON
	// {"<key>": "value"}
	// We expect NO redaction because "authority" is not sensitive.
	// If "auth" is detected at boundary, it will be redacted.
	input := `{"` + key + `": "value"}`

	redacted := RedactJSON([]byte(input))

	// If redacted, the value "value" will be replaced by "[REDACTED]"
	// Or rather, the whole JSON is reconstructed.

	// Check if "value" is present.
	if !strings.Contains(string(redacted), `"value"`) {
		t.Errorf("False positive redaction! 'authority' was mistaken for 'auth'.")
	}

	// Also verify that valid redaction still works
	// "auth": "value" -> redacted
	inputSensitive := `{"\nauth": "value"}`
	redactedSensitive := RedactJSON([]byte(inputSensitive))
	if strings.Contains(string(redactedSensitive), `"value"`) {
		t.Errorf("Failed to redact sensitive key '\\nauth'")
	}
}

func TestRedactSplitSensitiveKey(t *testing.T) {
    // Verify that a sensitive key split across chunks IS detected.
    // "authorization" is 13 chars.
    // Overlap is 64.
    // We want "authorization" to cross the 4096 boundary.
    // e.g. "authori" ends at 4095, "zation" starts at 4096.

    // Save original limit
	originalLimit := maxUnescapeLimit
	defer func() { maxUnescapeLimit = originalLimit }()
	maxUnescapeLimit = 100

    // Unescaped:
    // prefix (4096 - 7 = 4089 bytes)
    // "authorization"

    sb := strings.Builder{}
	sb.WriteString(`\n`)
	for i := 0; i < 4088; i++ { // 4089 - 1 (for \n) = 4088
		sb.WriteByte('x')
	}
	sb.WriteString("authorization")

    key := sb.String()
    input := `{"` + key + `": "value"}`

    redacted := RedactJSON([]byte(input))

    // Should be redacted
     if strings.Contains(string(redacted), `"value"`) {
		t.Errorf("Failed to redact split sensitive key 'authorization'")
	}
}
