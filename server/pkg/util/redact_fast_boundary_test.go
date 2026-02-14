// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRedactFast_Boundary tests the buffer boundary conditions in scanEscapedKeyForSensitive.
// The internal buffer size is 4097 bytes, and flushing occurs when bufIdx reaches 4096.
func TestRedactFast_Boundary(t *testing.T) {
	// The internal buffer size in scanEscapedKeyForSensitive is 4097.
	// Flushing logic triggers when bufIdx == bufSize-1 (4096).
	const flushThreshold = 4096

	t.Run("sensitive word at exact buffer boundary", func(t *testing.T) {
		// Construct a key where "password" ends exactly at the flush threshold.
		// "password" is 8 chars.
		// We need 4096 - 8 = 4088 padding characters.

		// To force scanEscapedKeyForSensitive, we need an escape sequence somewhere.
		// Let's escape the first char.
		escapedKey := "\\u0061" + strings.Repeat("a", flushThreshold-8-1) + "password"

		input := `{"` + escapedKey + `": "value"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]", "Should redact when sensitive word is at buffer boundary")
	})

	t.Run("sensitive word straddling buffer boundary", func(t *testing.T) {
		// "pass" at end of buffer, "word" at start of next chunk.
		// Padding needed: 4096 - 4 = 4092.

		// We escape the first char to force the complex path.
		escapedKey := "\\u0062" + strings.Repeat("b", flushThreshold-4-1) + "password"

		// Verify length logic (approximate due to escape expansion):
		// Unescaped length: 1 (from \u0062) + (4091) + 8 = 4100.
		// The buffer fills up to 4096.
		// "pass" is at indices 4092, 4093, 4094, 4095.
		// "word" is at indices 4096, 4097, 4098, 4099.
		// The overlap logic (64 bytes) should catch this.

		input := `{"` + escapedKey + `": "value"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]", "Should redact when sensitive word straddles buffer boundary")
	})

	t.Run("escape sequence straddling buffer boundary", func(t *testing.T) {
		// Construct a key where `\u0070` (p) is split across the boundary.
		// We want the escape sequence to start just before the flush threshold.
		// \u0070 is 6 chars.
		// If we start at 4096 - 3, we have `\u0` at end of buffer, `070` at start of next.
		// The unescape logic handles this by reading from the input stream directly,
		// so it shouldn't be affected by the output buffer flush,
		// BUT the *output* buffer might get flushed in the middle of writing the unescaped char?
		// No, `buf[bufIdx] = c` happens after unescaping.

		// However, if the unescaped char (p) lands at the boundary, and it's part of "password"...

		// Let's try to split the *input* escape sequence.
		// The loop `for i < n` iterates over input.
		// The buffer `buf` stores unescaped chars.

		// If we have "pass" + "\u0077" (w) + "ord".
		// And "\u0077" is processed when buffer is full?
		// No, buffer fills with *unescaped* chars.

		// Scenario: Buffer is at 4095. Next char is 'w'.
		// 1. Unescape '\u0077' -> 'w'.
		// 2. buf[4095] = 'w'. bufIdx becomes 4096.
		// 3. Flush condition met.
		// 4. "password" is now "passw" in buffer (split).
		// 5. "ord" comes in next chunk.
		// 6. Overlap logic should handle "ssw" + "ord".

		// ends at 4091
		// "pass" (4 chars) -> buffer 4091..4095 (wait, 4091+4=4095. One more needed to hit 4096)
		// Let's use exact counting.
		// We want "passw" to fill the buffer.
		// "passw" is 5 chars.
		// Padding needed: 4096 - 5 = 4091.

		escapedKey := "\\u0063" + strings.Repeat("c", flushThreshold-5-1) + "password"
		// "password" starts at index 4091.
		// p(4091), a(4092), s(4093), s(4094), w(4095).
		// bufIdx becomes 4096. Flush!
		// "passw" is in the chunk.
		// "ord" is in the next chunk.
		// Overlap should catch it.

		input := `{"` + escapedKey + `": "value"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]", "Should redact when sensitive word is split by buffer flush")
	})

	t.Run("invalid escape sequence at boundary", func(t *testing.T) {
		// Test handling of invalid escape at the boundary.
		// `\z` (invalid) at the end of buffer.
		// 4095 chars
		// Next chars: \z
		// The '\' is processed, i increments.
		// Then `switch keyContent[i]` sees 'z' (default case).
		// It writes 'z' to buf[4095].
		// bufIdx becomes 4096. Flush.

		// key ends with ...d\z

		// If the key also contains "password" before the invalid escape?
		// ...password\z
		// "password" ends at 4095. \z is at 4096.

		keyWithPassword := "\\u0064" + strings.Repeat("d", flushThreshold-8-1) + "password\\z"

		input := `{"` + keyWithPassword + `": "value"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]", "Should redact even with invalid escape at boundary")
	})

	t.Run("very long key with multiple flushes", func(t *testing.T) {
		// Key larger than 2 buffers (8KB+).
		// "password" in the middle of the second chunk.

		// Chunk 1: 4096 chars.
		// Chunk 2: 4096 chars. "password" is here.

		padding := strings.Repeat("e", flushThreshold + 100)
		escapedKey := "\\u0065" + padding[1:] + "password"

		input := `{"` + escapedKey + `": "value"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]", "Should redact in very long key")
	})
}
