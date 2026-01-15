// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_LongKeyWithEscape(t *testing.T) {
	// > 1024 bytes key with escape
	// We use "password" followed by non-alpha characters to ensure it matches "password" sensitive key
	// without being treated as a prefix of a longer word (like "passwords").

	padding := strings.Repeat("1", 1024)
	longKey := "password" + padding + "\\n" // escape sequence to trigger the branch

	input := `{"` + longKey + `": "sensitive"}`
	result := RedactJSON([]byte(input))

	// IsSensitiveKey should find "password" at the start.
	// So it should redact.
	assert.Contains(t, string(result), `"[REDACTED]"`)
}

func TestRedactJSON_LongKeySensitive_SIMD(t *testing.T) {
	// Trigger the > 128 bytes path in scanForSensitiveKeys
	// We need a key > 128 bytes that contains a sensitive word.
	// Use non-alpha padding to avoid word boundary issues.

	padding := strings.Repeat("1", 200)
	longKey := padding + "password"

	input := `{"` + longKey + `": "sensitive"}`
	result := RedactJSON([]byte(input))

	assert.Contains(t, string(result), `"[REDACTED]"`)
}

func TestRedactJSON_LongKeyNotSensitive_SIMD(t *testing.T) {
	// Trigger the > 128 bytes path in scanForSensitiveKeys but NO sensitive match

	padding := strings.Repeat("a", 200)
	longKey := padding + "safe"

	input := `{"` + longKey + `": "value"}`
	result := RedactJSON([]byte(input))

	assert.Contains(t, string(result), `"value"`)
}

