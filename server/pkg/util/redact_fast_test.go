// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestRedactJSONFast(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 1. Happy Path
		{
			name:     "simple sensitive key",
			input:    `{"password": "secret"}`,
			expected: `{"password": "[REDACTED]"}`,
		},
		{
			name:     "multiple sensitive keys",
			input:    `{"user": "admin", "password": "123", "token": "abc"}`,
			expected: `{"user": "admin", "password": "[REDACTED]", "token": "[REDACTED]"}`,
		},
		{
			name:     "non-sensitive key",
			input:    `{"public": "data"}`,
			expected: `{"public": "data"}`,
		},

		// 2. Nested Structures
		{
			name:     "nested object sensitive",
			input:    `{"config": {"password": "secret"}}`,
			expected: `{"config": {"password": "[REDACTED]"}}`,
		},
		{
			name:     "array of objects sensitive",
			input:    `{"users": [{"id": 1, "token": "abc"}, {"id": 2, "token": "def"}]}`,
			expected: `{"users": [{"id": 1, "token": "[REDACTED]"}, {"id": 2, "token": "[REDACTED]"}]}`,
		},
		{
			name:     "nested mixed",
			input:    `{"data": {"nested": [{"val": 1}, {"secret": "hide"}]}}`,
			expected: `{"data": {"nested": [{"val": 1}, {"secret": "[REDACTED]"}]}}`,
		},

		// 3. Edge Cases
		{
			name:     "empty object",
			input:    `{}`,
			expected: `{}`,
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: `[]`,
		},
		{
			name:     "null value",
			input:    `{"key": null}`,
			expected: `{"key": null}`,
		},
		{
			name:     "boolean values",
			input:    `{"key": true, "other": false}`,
			expected: `{"key": true, "other": false}`,
		},
		{
			name:     "number values",
			input:    `{"id": 123, "cost": 12.34}`,
			expected: `{"id": 123, "cost": 12.34}`,
		},
		{
			name:     "sensitive key with null value",
			input:    `{"password": null}`,
			expected: `{"password": "[REDACTED]"}`,
		},
		{
			name:     "sensitive key with number value",
			input:    `{"token": 12345}`,
			expected: `{"token": "[REDACTED]"}`,
		},
		{
			name:     "sensitive key with boolean value",
			input:    `{"secret_flag": true}`,
			expected: `{"secret_flag": "[REDACTED]"}`,
		},
		{
			name:     "sensitive key with array value",
			input:    `{"secrets": ["a", "b"]}`,
			expected: `{"secrets": "[REDACTED]"}`,
		},
		{
			name:     "sensitive key with object value",
			input:    `{"credentials": {"user": "a"}}`,
			expected: `{"credentials": "[REDACTED]"}`,
		},

		// 4. Whitespace
		{
			name:     "whitespace handling",
			input:    `{  "password"  :   "secret"  }`,
			expected: `{  "password"  :   "[REDACTED]"  }`,
		},
		{
			name:     "newlines and tabs",
			input:    "{\n\t\"password\":\n\t\"secret\"\n}",
			expected: "{\n\t\"password\":\n\t\"[REDACTED]\"\n}",
		},

		// 5. Malformed JSON (Robustness)
		// Note: The function does not validate JSON, it just scans.
		// If it looks like a key-value pair, it might redact it.
		// If it's totally garbage, it should return as is.
		{
			name:     "malformed unclosed brace",
			input:    `{"password": "secret"`,
			expected: `{"password": "[REDACTED]"`, // It should still find and redact the key
		},
		{
			name:     "malformed unclosed string",
			input:    `{"password": "sec`,
			expected: `{"password": "[REDACTED]"`, // Safer behavior: redact partial value if we think it's sensitive
		},
		{
			name:     "malformed garbage",
			input:    `not json`,
			expected: `not json`,
		},
		{
			name:     "malformed just a key",
			input:    `"password": "secret"`,
			expected: `"password": "[REDACTED]"`, // Should work even without braces if the structure matches key:value
		},

		// 6. Escaped Characters
		{
			name:     "escaped quote in key",
			input:    `{"pass\"word": "secret"}`,
			expected: `{"pass\"word": "secret"}`, // "pass\"word" is not "password"
		},
		{
			name:     "escaped backslash in key (not sensitive)",
			input:    `{"pass\\word": "secret"}`,
			expected: `{"pass\\word": "secret"}`,
		},
		{
			name:     "escaped char in key (unicode)",
			input:    `{"p\u0061ssword": "secret"}`,
			expected: `{"p\u0061ssword": "[REDACTED]"}`,
		},
		{
			name:     "escaped quote in value",
			input:    `{"key": "val\"ue"}`,
			expected: `{"key": "val\"ue"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactJSONFast([]byte(tt.input))
			// Normalize JSON for comparison where possible, or string compare for robustness/whitespace preservation
			// Since redactJSONFast preserves whitespace, exact string match is preferred if expected is exact.
			if string(got) != tt.expected {
				t.Errorf("redactJSONFast() = %q, want %q", string(got), tt.expected)
			}
		})
	}
}

func TestRedactJSONFast_LargeBuffer(t *testing.T) {
	// Generate a large JSON to trigger buffer resizing/lazy allocation logic
	// We want enough data to potentially exceed initial buffer estimates
	count := 5000
	var sb strings.Builder
	sb.WriteString(`{"data": [`)
	for i := 0; i < count; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf(`{"id": %d, "password": "secret-%d", "safe": "value-%d"}`, i, i, i))
	}
	sb.WriteString(`]}`)

	input := []byte(sb.String())
	got := redactJSONFast(input)

	// Verify redaction happened
	if bytes.Contains(got, []byte(`"secret-0"`)) {
		t.Error("Failed to redact sensitive data in large input")
	}
	if !bytes.Contains(got, []byte(`[REDACTED]`)) {
		t.Error("Redacted placeholder not found in large input")
	}
	if !bytes.Contains(got, []byte(`"value-0"`)) {
		t.Error("Non-sensitive data lost in large input")
	}

	// Verify valid JSON structure
	var result map[string]interface{}
	if err := json.Unmarshal(got, &result); err != nil {
		t.Errorf("Result is not valid JSON: %v", err)
	}
}

func TestRedactJSONFast_DeepNesting(t *testing.T) {
	// Verify that deep nesting doesn't cause stack overflow or logic errors
	// (though the implementation is iterative, so stack overflow is unlikely)
	depth := 100
	var inputBuilder, expectedBuilder strings.Builder

	for i := 0; i < depth; i++ {
		inputBuilder.WriteString(`{"nested": `)
		expectedBuilder.WriteString(`{"nested": `)
	}

	inputBuilder.WriteString(`{"password": "deep_secret"}`)
	expectedBuilder.WriteString(`{"password": "[REDACTED]"}`)

	for i := 0; i < depth; i++ {
		inputBuilder.WriteString(`}`)
		expectedBuilder.WriteString(`}`)
	}

	got := redactJSONFast([]byte(inputBuilder.String()))
	if string(got) != expectedBuilder.String() {
		t.Errorf("Deep nesting failed.\nGot: %s\nWant: %s", string(got), expectedBuilder.String())
	}
}
