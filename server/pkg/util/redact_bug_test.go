// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_EdgeCases(t *testing.T) {
	// Bug hypothesis: `redactJSONFast` might fail if the sensitive key is at the very end of the JSON object,
	// or if the value contains characters that confuse the scanner (like braces or brackets in strings).

	// Case 1: Value contains characters that look like JSON structure
	t.Run("value with braces", func(t *testing.T) {
		input := `{"password": "{not an object}"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
		assert.NotContains(t, string(output), "{not an object}")
	})

	// Case 2: Value with escaped quotes
	t.Run("value with escaped quotes", func(t *testing.T) {
		input := `{"password": "pass\"word"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
		// "password" contains "pass", so checking NotContains "pass" fails on the key itself.
		// We should check that the value is redacted.
		var m map[string]interface{}
		err := json.Unmarshal(output, &m)
		assert.NoError(t, err)
		assert.Equal(t, "[REDACTED]", m["password"])
	})

	// Case 3: Key with weird spacing
	t.Run("key with spacing", func(t *testing.T) {
		input := `{"password"   :    "secret"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
	})

	// Case 4: Multiple sensitive keys
	t.Run("multiple sensitive keys", func(t *testing.T) {
		input := `{"password": "p1", "token": "t1"}`
		output := RedactJSON([]byte(input))
		// Check counts
		str := string(output)

		assert.Contains(t, str, "[REDACTED]")
		assert.NotContains(t, str, "p1")
		assert.NotContains(t, str, "t1")

		var m map[string]interface{}
		json.Unmarshal(output, &m)
		assert.Equal(t, "[REDACTED]", m["password"])
		assert.Equal(t, "[REDACTED]", m["token"])
	})

	// Case 5: Nested sensitive key inside a string value (should NOT be redacted if it's just a substring of value)
	// But wait, our scanner `scanJSONForSensitiveKeys` is smart about keys.
	// But `redactJSONFast` iterates through all strings.
	t.Run("sensitive word in value", func(t *testing.T) {
		input := `{"description": "this contains password word"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "this contains password word")
	})

	// Case 6: Malformed JSON - should return original
	t.Run("malformed json unclosed string", func(t *testing.T) {
		input := `{"password": "unclosed`
		output := RedactJSON([]byte(input))
		assert.Equal(t, input, string(output))
	})

	// Case 7: Malformed JSON - no colon
	t.Run("malformed json no colon", func(t *testing.T) {
		input := `{"password" "secret"}`
		output := RedactJSON([]byte(input))
		// Depending on implementation, it might redact if it finds "password" string, but `redactJSONFast` looks for colon.
		// `redactJSONFast` implementation:
		// `isKey` check loops for colon. If it hits another quote (start of "secret"), it might fail to see colon?
		// Let's check `redactJSONFast` code.
		// It loops `j < n`. It breaks if it sees something that is not whitespace.
		// If it sees '"', it breaks and `isKey` is false.
		// So it should NOT redact.
		assert.Equal(t, input, string(output))
	})

	// Case 8: Key matches prefix of sensitive key but is not sensitive
	// e.g. "authentication" starts with "auth", but "authentication" is not in the list, "auth" is.
	// `checkPotentialMatch` logic handles boundaries.
	t.Run("prefix match false positive", func(t *testing.T) {
		input := `{"authentication": "visible"}`
		// "auth" is sensitive. "authentication" starts with "auth".
		// `checkPotentialMatch` checks boundary.
		// "authentication" has 'e' after "auth". 'e' is lowercase.
		// "auth" key ends at index 4. input[4] is 'e'.
		// Logic: `if next >= 'a' && next <= 'z' { continue }`
		// So it should continue and NOT match "auth".
		// Then it checks other keys.
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "visible")
	})

	// Case 9: JSON string array with sensitive-looking strings
	t.Run("sensitive string in array", func(t *testing.T) {
		input := `["password", "token"]`
		output := RedactJSON([]byte(input))
		// Should NOT redact because they are values, not keys
		assert.Contains(t, string(output), "password")
		assert.Contains(t, string(output), "token")
	})

	// Case 10: JSON object where value is a number 0
	t.Run("zero value", func(t *testing.T) {
		input := `{"password": 0}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
	})

	// Case 11: JSON object where value is empty object
	t.Run("empty object value", func(t *testing.T) {
		input := `{"password": {}}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
		assert.NotContains(t, string(output), "{}")
	})

	// Case 12: Empty key
	t.Run("empty key", func(t *testing.T) {
		input := `{"": "val"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "val")
	})

	// Case 13: Uppercase key with mixed casing
	t.Run("mixed case key", func(t *testing.T) {
		// "auth" is in the list.
		// "Auth" should match.
		input := `{"Auth": "secret"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
	})
}
