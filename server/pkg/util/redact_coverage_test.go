// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactCoverage_LargeKey(t *testing.T) {
	// Trigger the >128 bytes path in scanForSensitiveKeys
	// Create a key that is 200 bytes long and contains "api_key"
	prefix := strings.Repeat("a", 150)
	key := prefix + "api_key"

	assert.True(t, IsSensitiveKey(key))

	// Also test via JSON redaction
	input := `{"` + key + `": "secret"}`
	output := RedactJSON([]byte(input))
	assert.Contains(t, string(output), `[REDACTED]`)
}

func TestRedactCoverage_LargeKeyWithEscape(t *testing.T) {
	// Trigger the >1024 bytes path in redactJSONFast for escaped keys
	// Key must contain backslash
	prefix := strings.Repeat("a", 1100)
	key := prefix + "api_key" + `\\` // End with backslash to trigger escape check

	input := `{"` + key + `": "secret"}`
	output := RedactJSON([]byte(input))
	assert.Contains(t, string(output), `[REDACTED]`)
}

func TestRedactCoverage_UnmarshalFailure(t *testing.T) {
	// Trigger the unmarshal failure path in redactJSONFast
	// Key has invalid escape sequence
	key := `api_key\u12` // Invalid unicode escape

	input := `{"` + key + `": "secret"}`
	output := RedactJSON([]byte(input))

	// It should fall back to raw key check.
	// Raw key is `api_key\u12`. Contains `api_key`. Should redact.
	assert.Contains(t, string(output), `[REDACTED]`)
}

func TestRedactCoverage_ShortStringOptimization(t *testing.T) {
	// Test the short string optimization (<128 bytes) in scanForSensitiveKeys
	// specifically checking potential matches

	// "auth" is sensitive. "author" is not.
	// "author" < 128 bytes.
	assert.False(t, IsSensitiveKey("author"))

	// "api_key" is sensitive.
	assert.True(t, IsSensitiveKey("my_api_key"))
}

func TestRedactCoverage_SkipObjectArray(t *testing.T) {
	// Exercise skipObject and skipArray fully

	// Nested object with string containing braces
	input := `{"api_key": "secret", "public": {"a": "{", "b": "}"}}`
	output := RedactJSON([]byte(input))
	assert.Contains(t, string(output), `[REDACTED]`)
	assert.Contains(t, string(output), `"public": {"a": "{", "b": "}"}`)

	// Nested array with string containing brackets
	input2 := `{"api_key": "secret", "public": ["[", "]"]}`
	output2 := RedactJSON([]byte(input2))
	assert.Contains(t, string(output2), `[REDACTED]`)
	assert.Contains(t, string(output2), `"public": ["[", "]"]`)
}

func TestRedactCoverage_UnclosedObjectArray(t *testing.T) {
	// Unclosed object
	input := `{"api_key": "secret", "public": {`
	output := RedactJSON([]byte(input))
	assert.Contains(t, string(output), `[REDACTED]`)
	// public value extends to end of string? No, RedactJSON returns input if no redaction happens.
	// But here redaction happens for api_key.
	// public value starts at {, calls skipObject.
	// skipObject scans until EOF. Returns len(input).
	// So public value is considered to be `{`.
	// Correct.

	// Unclosed array
	input2 := `{"api_key": "secret", "public": [`
	output2 := RedactJSON([]byte(input2))
	assert.Contains(t, string(output2), `[REDACTED]`)
}

func TestRedactCoverage_RedactMapDeep(t *testing.T) {
	// deeply nested map to ensure recursion works
	m := map[string]interface{}{
		"l1": map[string]interface{}{
			"l2": map[string]interface{}{
				"api_key": "secret",
			},
		},
	}
	redacted := RedactMap(m)
	l1 := redacted["l1"].(map[string]interface{})
	l2 := l1["l2"].(map[string]interface{})
	assert.Equal(t, "[REDACTED]", l2["api_key"])
}

func TestRedactCoverage_ScanSensitiveKeys(t *testing.T) {
	// Trigger both idxL and idxU paths in scanForSensitiveKeys (long string)

	// Need > 128 bytes.
	// "api_key" starts with 'a'.
	// "API_KEY" starts with 'A'.

	// Case 1: 'a' found first.
	s1 := strings.Repeat("x", 130) + "api_key"
	assert.True(t, IsSensitiveKey(s1))

	// Case 2: 'A' found first.
	s2 := strings.Repeat("x", 130) + "API_KEY"
	assert.True(t, IsSensitiveKey(s2))

	// Case 3: 'a' and 'A' both present, 'a' first.
	s3 := strings.Repeat("x", 130) + "api_key API_KEY"
	assert.True(t, IsSensitiveKey(s3))

    // Case 4: 'a' and 'A' both present, 'A' first.
	s4 := strings.Repeat("x", 130) + "API_KEY api_key"
	assert.True(t, IsSensitiveKey(s4))

	// Case 5: No match
	s5 := strings.Repeat("x", 130) + "nothing"
	assert.False(t, IsSensitiveKey(s5))
}

func TestRedactCoverage_NextCharOptimization(t *testing.T) {
    // "abacus". 'a' is start char. 'b' is next.
    // 'b' is not a second char for any sensitive key starting with 'a'.
    // (api_key -> p, auth -> u, etc.)
    // sensitiveKeyGroups['a'] = {api_key, apikey, auth, authorization...}
    // sensitiveNextCharMask['a'] has bits for p, u.
    // 'b' bit is 0.
    // So it should return false early.
    assert.False(t, IsSensitiveKey("abacus"))
}

func TestRedactCoverage_EOFWhitespace(t *testing.T) {
    // redactJSONFast loop end
    // JSON ending with whitespace after string
    input := `{"key": "val"   `
    // malformed but loop handles it
    output := RedactJSON([]byte(input))
    // Should be input
    assert.Equal(t, input, string(output))
}
