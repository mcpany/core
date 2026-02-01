// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_Intervention(t *testing.T) {
	t.Parallel()

	// 1. Deep Nesting
	t.Run("Deep Nesting", func(t *testing.T) {
		depth := 2000
		var sb strings.Builder
		for i := 0; i < depth; i++ {
			sb.WriteString(`{"a":`)
		}
		sb.WriteString(`{"password": "secret"}`)
		for i := 0; i < depth; i++ {
			sb.WriteString("}")
		}

		input := []byte(sb.String())
		output := RedactJSON(input)

		// We expect the "secret" to be redacted
		assert.Contains(t, string(output), `"[REDACTED]"`)
		// And the structure should be preserved (length check is a proxy)
		// "secret" -> "[REDACTED]" is +4 chars
		// But wait, "secret" is 6 chars, "[REDACTED]" is 10 chars. So +4.
		// Original length: depth*5 + 22 + depth
		// New length should be Original + 4.
		assert.Equal(t, len(input)+4, len(output))
	})

	// 2. Large Key (Streaming Scanner)
	t.Run("Large Key", func(t *testing.T) {
		// Create a key larger than maxUnescapeLimit (1MB default) to exercise the streaming path.
		// Although a 2MB key composed of 'a's will not match any sensitive key (which are short constants),
		// this test ensures the scanner handles large inputs without crashing or excessive allocation.
		size := 2 * 1024 * 1024
		longKey := strings.Repeat("a", size)
		input := `{"` + longKey + `": "value"}`

		output := RedactJSON([]byte(input))
		assert.Equal(t, input, string(output))
	})

	// 3. Comment Edge Cases
	commentTests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Comment between key and colon",
			input:    `{"password" /* comment */ : "secret"}`,
			expected: `{"password" /* comment */ : "[REDACTED]"}`,
		},
		{
			name:     "Comment between colon and value",
			input:    `{"password": /* comment */ "secret"}`,
			expected: `{"password": /* comment */ "[REDACTED]"}`,
		},
		{
			name:     "Comment inside value (not a comment)",
			input:    `{"key": "http://example.com"}`,
			expected: `{"key": "http://example.com"}`,
		},
		{
			name:     "Slash not comment",
			input:    `{"password": "a/b"}`,
			expected: `{"password": "[REDACTED]"}`,
		},
	}

	for _, tt := range commentTests {
		t.Run(tt.name, func(t *testing.T) {
			output := RedactJSON([]byte(tt.input))
			assert.Equal(t, tt.expected, string(output))
		})
	}

	// 4. Escape Sequences
	escapeTests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Unicode escape in key",
			input:    `{"p\u0061ssword": "secret"}`, // \u0061 is 'a'
			expected: `{"p\u0061ssword": "[REDACTED]"}`,
		},
		{
			name:     "Escaped quote in key",
			input:    `{"pass\"word": "value"}`, // key is pass"word, not sensitive
			expected: `{"pass\"word": "value"}`,
		},
		{
			name:     "Escaped backslash in key",
			input:    `{"pass\\word": "value"}`, // key is pass\word
			expected: `{"pass\\word": "value"}`,
		},
		{
			name:     "Null byte escape",
			input:    `{"pass\u0000word": "value"}`,
			expected: `{"pass\u0000word": "value"}`,
		},
	}

	for _, tt := range escapeTests {
		t.Run(tt.name, func(t *testing.T) {
			output := RedactJSON([]byte(tt.input))
			assert.Equal(t, tt.expected, string(output))
		})
	}

	// 5. Malformed Inputs
	malformedTests := []struct {
		name  string
		input string
	}{
		{"Truncated after key", `{"password"`},
		{"Truncated after colon", `{"password":`},
		{"Truncated inside string", `{"password": "sec`},
		{"Unclosed quote in key", `{"password : "val"}`},
		{"Comment unclosed", `{"key": "val" /* unclosed `},
	}

	for _, tt := range malformedTests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not crash or hang
			output := RedactJSON([]byte(tt.input))
			// Output should roughly match input (maybe exact, maybe redacted if it parsed enough)
			// For truncated inputs, we generally expect return-as-is or partial.
			// The main goal here is NO PANIC.
			assert.NotNil(t, output)
		})
	}
}
