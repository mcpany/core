// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSONFast(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Base Cases
		{
			name:     "empty json",
			input:    `{}`,
			expected: `{}`,
		},
		{
			name:     "simple object",
			input:    `{"name": "test"}`,
			expected: `{"name": "test"}`,
		},
		{
			name:     "simple sensitive",
			input:    `{"password": "secret"}`,
			expected: `{"password": "[REDACTED]"}`,
		},

		// Escaping
		{
			name:     "escaped quote in value",
			input:    `{"key": "val\"ue"}`,
			expected: `{"key": "val\"ue"}`,
		},
		{
			name:     "escaped quote in key",
			input:    `{"ke\"y": "value"}`,
			expected: `{"ke\"y": "value"}`,
		},
		{
			name:     "escaped backslash",
			input:    `{"key": "val\\ue"}`,
			expected: `{"key": "val\\ue"}`,
		},
		{
			name:     "escaped quote at end of value",
			input:    `{"key": "value\""}`,
			expected: `{"key": "value\""}`,
		},
		{
			name:     "multiple escaped backslashes",
			input:    `{"key": "\\\\"}`, // key: \\
			expected: `{"key": "\\\\"}`,
		},
		{
			name:     "escaped quote preceded by escaped backslash",
			input:    `{"key": "\\\""}`, // value is \"
			expected: `{"key": "\\\""}`,
		},
		{
			name:     "escaped quote preceded by double escaped backslash",
			input:    `{"key": "\\\\\""}`, // value is \\" -> quote is NOT escaped
			expected: `{"key": "\\\\\""}`,
		},
        {
            name:     "sensitive key with escaped quote",
            input:    `{"pass\"word": "secret"}`, // key: pass"word -> should NOT match "password"
            expected: `{"pass\"word": "secret"}`,
        },
        {
            name:     "sensitive key with escaped chars",
            input:    `{"\u0070assword": "secret"}`, // \u0070 -> p
            expected: `{"\u0070assword": "[REDACTED]"}`,
        },

		// Unicode
		{
			name:     "unicode value",
			input:    `{"key": "\u263A"}`,
			expected: `{"key": "\u263A"}`,
		},
		{
			name:     "unicode sensitive key",
			input:    `{"\u0070assword": "secret"}`,
			expected: `{"\u0070assword": "[REDACTED]"}`,
		},
        {
            name: "unicode sensitive key mixed",
            input: `{"p\u0061ssword": "secret"}`, // p\u0061ssword -> password
            expected: `{"p\u0061ssword": "[REDACTED]"}`,
        },

		// Comments
		{
			name:     "line comment before key",
			input:    `{ // comment
"password": "secret"}`,
			expected: `{ // comment
"password": "[REDACTED]"}`,
		},
		{
			name:     "block comment before key",
			input:    `{ /* comment */ "password": "secret"}`,
			expected: `{ /* comment */ "password": "[REDACTED]"}`,
		},
		{
			name:     "comment between key and colon",
			input:    `{"password" /* comment */ : "secret"}`,
			expected: `{"password" /* comment */ : "[REDACTED]"}`,
		},
		{
			name:     "comment between colon and value",
			input:    `{"password": /* comment */ "secret"}`,
			expected: `{"password": /* comment */ "[REDACTED]"}`,
		},
		{
			name:     "comment inside array",
			input:    `{"key": [1, /* comment */ 2]}`,
			expected: `{"key": [1, /* comment */ 2]}`,
		},
        {
            name:     "fake comment in string",
            input:    `{"key": "http://example.com"}`,
            expected: `{"key": "http://example.com"}`,
        },
        {
            name:     "fake block comment in string",
            input:    `{"key": "/* not a comment */"}`,
            expected: `{"key": "/* not a comment */"}`,
        },

		// Whitespace
		{
			name:     "whitespace everywhere",
			input:    `{  "password"  :  "secret"  }`,
			expected: `{  "password"  :  "[REDACTED]"  }`,
		},
		{
			name:     "tabs and newlines",
			input:    "{\n\t\"password\":\n\t\"secret\"\n}",
			expected: "{\n\t\"password\":\n\t\"[REDACTED]\"\n}",
		},

		// Data Types
		{
			name:     "sensitive number",
			input:    `{"password": 12345}`,
			expected: `{"password": "[REDACTED]"}`,
		},
		{
			name:     "sensitive boolean",
			input:    `{"password": true}`,
			expected: `{"password": "[REDACTED]"}`,
		},
		{
			name:     "sensitive null",
			input:    `{"password": null}`,
			expected: `{"password": "[REDACTED]"}`,
		},
		{
			name:     "scientific notation",
			input:    `{"password": 1.23e+10}`,
			expected: `{"password": "[REDACTED]"}`,
		},
        {
            name:     "negative number",
            input:    `{"password": -123}`,
            expected: `{"password": "[REDACTED]"}`,
        },

		// Nesting
		{
			name:     "nested object",
			input:    `{"user": {"password": "secret"}}`,
			expected: `{"user": {"password": "[REDACTED]"}}`,
		},
		{
			name:     "nested array",
			input:    `{"users": [{"password": "s1"}, {"password": "s2"}]}`,
			expected: `{"users": [{"password": "[REDACTED]"}, {"password": "[REDACTED]"}]}`,
		},
        {
            name:     "deeply nested",
            input:    `{"a": {"b": {"c": {"password": "secret"}}}}`,
            expected: `{"a": {"b": {"c": {"password": "[REDACTED]"}}}}`,
        },

		// Malformed Inputs (Sanity Checks - should not panic/hang)
		{
			name:     "unclosed string",
			input:    `{"key": "value`,
			expected: `{"key": "value`,
		},
		{
			name:     "unclosed object",
			input:    `{"key": "value"`,
			expected: `{"key": "value"`,
		},
        {
            name:     "colon without value",
            input:    `{"password": }`,
            expected: `{"password": "[REDACTED]"}`,
        },
        {
            name:     "key without colon",
            input:    `{"password" "secret"}`,
            expected: `{"password" "secret"}`,
        },

        // Edge Cases
        {
            name:     "case sensitivity", // "password" is sensitive, "PASSWORD" might be depending on config.
            // In redact.go, keys are lowercase. But checkPotentialMatch handles upper case logic.
            // Test "Password" -> likely sensitive.
            input:    `{"Password": "secret"}`,
            expected: `{"Password": "[REDACTED]"}`,
        },
        {
            name:     "partial match prefix", // "auth" is sensitive. "author" should NOT be.
            input:    `{"author": "name"}`,
            expected: `{"author": "name"}`,
        },
        {
            name:     "partial match suffix", // "word" is NOT sensitive. "password" is.
            input:    `{"myword": "val"}`,
            expected: `{"myword": "val"}`,
        },
        {
            name:     "snake case match",
            input:    `{"api_key": "123"}`,
            expected: `{"api_key": "[REDACTED]"}`,
        },

        // Large Inputs / Limits
        {
            name: "large ignored value",
            input: `{"key": "` + strings.Repeat("a", 10000) + `"}`,
            expected: `{"key": "` + strings.Repeat("a", 10000) + `"}`,
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Because redactJSONFast logic for null might be tricky (it appends redactedValue bytes),
			// and redactedValue is defined in redact.go as `"[REDACTED]"` (a JSON string).
			// If input is `{"p": null}`, output is `{"p": "[REDACTED]"}`.
			result := redactJSONFast([]byte(tt.input))
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

// TestUnescapeKeySmall tests the internal unescapeKeySmall function
func TestUnescapeKeySmall(t *testing.T) {
    // This requires a buffer
    buf := make([]byte, 1024)

    tests := []struct {
        name     string
        input    string
        expected string
        ok       bool
    }{
        {"no escape", "abc", "abc", true},
        {"simple escape", `a\nb`, "a\nb", true},
        {"unicode escape", `\u0061`, "a", true},
        {"unicode escape high", `\u263A`, "?", true}, // ☺ -> replaced by ? as per implementation
        // Note: unescapeKeySmall output depends on implementation details (utf8 encoding vs byte values).
        // The implementation creates a byte slice. For \u0061 it writes 'a'.
        // For \u263A, it writes the utf8 bytes.

        {"escaped quote", `\"`, `"`, true},
        {"escaped backslash", `\\`, `\`, true},
        {"escaped slash", `\/`, `/`, true},
        {"escaped backspace", `\b`, "\b", true},
        {"escaped formfeed", `\f`, "\f", true},
        {"escaped carriage return", `\r`, "\r", true},
        {"escaped tab", `\t`, "\t", true},

        {"invalid escape", `\z`, "z", true}, // implementation falls back to char
        {"incomplete unicode", `\u006`, "u006", true}, // falls back to u and treats rest as literals
        {"malformed unicode", `\u006z`, "u006z", true}, // falls back to u and treats rest as literals
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            res, ok := unescapeKeySmall([]byte(tt.input), buf)
            if !tt.ok {
                assert.False(t, ok)
            } else {
                assert.True(t, ok)
                assert.Equal(t, tt.expected, string(res))
            }
        })
    }
}

func TestSkipJSONValue(t *testing.T) {
     tests := []struct {
        name     string
        input    string
        expected string // the remainder of string
    }{
        {"string", `"abc" def`, " def"},
        {"string escaped", `"a\"b" def`, " def"},
        {"number", `123 def`, " def"},
        {"number neg", `-123 def`, " def"},
        {"number float", `1.23 def`, " def"},
        {"number exp", `1e10 def`, " def"},
        {"true", `true def`, " def"},
        {"false", `false def`, " def"},
        {"null", `null def`, " def"},
        {"object", `{"a":1} def`, " def"},
        {"object nested", `{"a":{"b":2}} def`, " def"},
        {"array", `[1,2] def`, " def"},
        {"array nested", `[1,[2,3]] def`, " def"},
        {"string with comments", `"abc" /* comment */ def`, " /* comment */ def"}, // skipJSONValue only skips value, not trailing comments
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            idx := skipJSONValue([]byte(tt.input), 0)
            assert.Equal(t, tt.expected, string(tt.input[idx:]))
        })
    }
}

func TestRedactFastLimits(t *testing.T) {
    // Test large key exceeding stack limit but handled by unescape
    // unescapeStackLimit is 256

    // Construct a key that is 300 chars long and sensitive ("password" + junk)
    // Actually if it's too long, it might not match "password".
    // But if we have a sensitive key that is huge?
    // "password" is short.
    // The issue is unescaping.

    // Construct a key that needs unescaping and is large
    largeKey := strings.Repeat("a", 300) + "nonsensitive"
    // This key ends in nonsensitive, so it shouldn't be redacted.
    input := `{"` + largeKey + `": "val"}`
    expected := input
    result := redactJSONFast([]byte(input))
    assert.Equal(t, expected, string(result))

    // Construct a key that IS "password" but represented with many escapes?
    // "p\u0061ssword" ...
    // If we make it huge with escapes: "\u0070\u0061..."

    // \u0070 is 6 chars. 8 chars in password. 6*8 = 48 chars. Not huge.

    // What if we have a really long key that IS sensitive?
    // "private_key" is sensitive.
    // Is there any sensitive key that can be arbitrarily long? No, they are fixed list.

    // However, the unescape logic is triggered for ANY key with backslashes.
    // So if we have a huge key with backslashes, it exercises the heap allocation path (or streaming path).

    hugeKey := strings.Repeat(`\u0061`, 1000) // 6000 chars, unescapes to 1000 'a's.
    input2 := `{"` + hugeKey + `": "val"}`
    // Should not panic
    result2 := redactJSONFast([]byte(input2))
    assert.Equal(t, input2, string(result2))
}
