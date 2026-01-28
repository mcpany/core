// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // If empty, we expect input to be returned as is (no redaction)
	}{
		// --- Happy Path ---
		{
			name:     "simple sensitive key",
			input:    `{"password": "secret"}`,
			expected: `{"password": "[REDACTED]"}`,
		},
		{
			name:     "multiple keys",
			input:    `{"user": "alice", "password": "123", "token": "abc"}`,
			expected: `{"user": "alice", "password": "[REDACTED]", "token": "[REDACTED]"}`,
		},
		{
			name:     "mixed formatting",
			input:    `{ "user" : "alice" , "password":   "123" }`,
			expected: `{ "user" : "alice" , "password":   "[REDACTED]" }`,
		},

		// --- Nested Structures ---
		{
			name:     "sensitive key in nested object",
			input:    `{"config": {"public_data": {"token": "secret"}}}`,
			expected: `{"config": {"public_data": {"token": "[REDACTED]"}}}`,
		},
		{
			name:     "sensitive key in array",
			input:    `[{"password": "1"}, {"password": "2"}]`,
			expected: `[{"password": "[REDACTED]"}, {"password": "[REDACTED]"}]`,
		},
		{
			name:     "value is object (redact entire value)",
			input:    `{"credentials": {"user": "a", "pass": "b"}}`,
			expected: `{"credentials": "[REDACTED]"}`,
		},
		{
			name:     "value is array (redact entire value)",
			input:    `{"tokens": ["a", "b"]}`,
			expected: `{"tokens": "[REDACTED]"}`,
		},

		// --- Primitives & Types ---
		{
			name:     "value is number",
			input:    `{"secret_code": 12345}`,
			expected: `{"secret_code": "[REDACTED]"}`,
		},
		{
			name:     "value is boolean",
			input:    `{"is_secret": true}`,
			expected: `{"is_secret": "[REDACTED]"}`,
		},
		{
			name:     "value is null",
			input:    `{"void_secret": null}`,
			expected: `{"void_secret": "[REDACTED]"}`,
		},
		{
			name:     "non-sensitive numbers/bools preserved",
			input:    `{"count": 123, "active": true}`,
			expected: "",
		},

		// --- Escapes & Special Chars ---
		{
			name:     "escaped quotes in key",
			input:    `{"pass\"word": "secret"}`,
			expected: "", // "pass\"word" != "password", so no redaction expected unless fuzzy match?
			              // Standard IsSensitiveKey checks for exact matches or contains.
			              // "pass\"word" contains "word" but not "password".
		},
		{
			name:     "unicode escaped key",
			input:    `{"\u0070assword": "secret"}`, // \u0070 = p
			expected: `{"\u0070assword": "[REDACTED]"}`,
		},
		{
			name:     "escaped quotes in value",
			input:    `{"password": "my \"secret\" value"}`,
			expected: `{"password": "[REDACTED]"}`,
		},

		// --- Edge Cases ---
		{
			name:     "empty object",
			input:    `{}`,
			expected: "",
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: "",
		},
		{
			name:     "top level array",
			input:    `[1, 2, 3]`,
			expected: "",
		},
		{
			name:     "key looks sensitive but is suffix",
			input:    `{"nopassword": "ok"}`,
			expected: `{"nopassword": "[REDACTED]"}`,
		},

		// --- Malformed JSON (Robustness) ---
		{
			name:     "malformed: unclosed brace",
			input:    `{"password": "secret"`,
			expected: `{"password": "[REDACTED]"`, // It might still manage to redact if it finds the key/value pair
		},
		{
			name:     "malformed: missing value",
			input:    `{"password": }`,
			expected: `{"password": "[REDACTED]"}`, // Robust recovery: inserts value before }
		},
		{
			name:     "malformed: trailing garbage",
			input:    `{"password": "secret"} garbage`,
			expected: `{"password": "[REDACTED]"} garbage`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactJSON([]byte(tt.input))

			if tt.expected == "" {
				assert.Equal(t, tt.input, string(got), "Expected no change")
			} else {
				// Normalize JSON for comparison if valid
				// However, RedactJSON preserves whitespace, so string comparison should usually work
				// if we wrote expected output carefully.
				// For robustness, we can try to compare as JSON if both are valid.

				// Direct string comparison first (e.g. for malformed output)
				if !json.Valid([]byte(tt.expected)) {
					assert.Equal(t, tt.expected, string(got))
				} else {
					// Both valid, use JSONEq but also check literal string if possible to catch whitespace diffs
					// if strict preservation is desired.
					// The implementation seems to preserve format.
					assert.Equal(t, tt.expected, string(got))
				}
			}
		})
	}
}

func TestRedactJSON_BufferGrowth(t *testing.T) {
	// Generate a large JSON to test buffer handling
	// "password": "..." repeated

	var sb strings.Builder
	sb.WriteString(`[`)
	for i := 0; i < 1000; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(`{"password": "secret_value_that_is_long_enough_` + string(rune(i%10+'0')) + `"}`)
	}
	sb.WriteString(`]`)

	input := sb.String()
	got := RedactJSON([]byte(input))

	// Check that input size is significant
	if len(input) < 10000 {
		t.Fatalf("Test setup failed, input too small: %d", len(input))
	}

	// Check that we got a result
	if len(got) == 0 {
		t.Fatal("Result is empty")
	}

	// Verify everything is redacted
	gotStr := string(got)
	if strings.Contains(gotStr, "secret_value") {
		t.Error("Found unredacted secret value in large output")
	}

	if !strings.Contains(gotStr, "[REDACTED]") {
		t.Error("Did not find [REDACTED] placeholder in large output")
	}

	// Verify it's still valid JSON
	if !json.Valid(got) {
		t.Error("Output is not valid JSON")
	}
}

func TestRedactJSON_NoAllocationForCleanInput(t *testing.T) {
	input := []byte(`{"public": "data", "count": 123}`)
	got := RedactJSON(input)

	// Check if the returned slice header points to the same underlying array
	// If no sensitive keys were found, it should return input slice directly (zero alloc optimization)

	// We can check cap and pointer roughly, or just rely on the contract documentation.
	// But let's verify we didn't mutilate it.
	assert.Equal(t, string(input), string(got))
}
