// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkipJSONValue_HunterCoverage(t *testing.T) {
	// Test skipJSONValue with various inputs to ensure all branches are covered.
	tests := []struct {
		name     string
		input    string
		expected string // The rest of the string after skipping the value
	}{
		{"string", `"abc" def`, " def"},
		{"string escaped", `"a\"b" def`, " def"},
		{"object", `{"a": 1} def`, " def"},
		{"object nested", `{"a": {"b": 2}} def`, " def"},
		{"array", `[1, 2] def`, " def"},
		{"array nested", `[1, [2, 3]] def`, " def"},
		{"true", `true def`, " def"},
		{"false", `false def`, " def"},
		{"null", `null def`, " def"},
		{"number", `123 def`, " def"},
		{"number negative", `-123 def`, " def"},
		{"number float", `1.23 def`, " def"},
		{"number exp", `1e10 def`, " def"},
		{"number delimiter comma", `123,`, ","},
		{"number delimiter bracket", `123}`, "}"},
		{"number delimiter square bracket", `123]`, "]"},
		{"number EOF", `123`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := []byte(tt.input)
			idx := skipJSONValue(input, 0)
			rest := string(input[idx:])
			assert.Equal(t, tt.expected, rest)
		})
	}
}

func TestRedactJSON_EdgeCases_Hunter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"malformed unclosed string sensitive", `{"api_key": "val`, `[REDACTED]`},
		{"malformed unclosed string as key", `{"key`, `{"key`},
		{"key with only whitespace value", `{"api_key":   }`, `[REDACTED]`},
		{"key with escaped backslash", `{"api_key": "val\\ue"}`, `[REDACTED]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RedactJSON([]byte(tt.input))
			assert.Contains(t, string(output), tt.contains)
		})
	}
}

func TestRedactMap_NestedSliceTypes_Hunter(t *testing.T) {
	// deeply nested mix of maps and slices
	input := map[string]interface{}{
		"l1": []interface{}{
			[]interface{}{
				map[string]interface{}{
					"api_key": "secret",
				},
			},
		},
	}
	redacted := RedactMap(input)

	// Traverse to check
	l1 := redacted["l1"].([]interface{})
	l2 := l1[0].([]interface{})
	m := l2[0].(map[string]interface{})
	assert.Equal(t, "[REDACTED]", m["api_key"])
}

func TestIsKey_Coverage_Hunter(t *testing.T) {
	// Case: EOF after closing quote
	// `isKey` returns false.
	// `redactJSONFast` continues.
	input2 := `{"key" `
	output := RedactJSON([]byte(input2))
	assert.Equal(t, input2, string(output))

	// Case: Escape in key
	// `isKey` conservatively returns true if escape found.
	input3 := `{"api_key\u0041": "val"}` // api_keyA -> Sensitive because of boundary check logic
	output3 := RedactJSON([]byte(input3))
	assert.Contains(t, string(output3), "[REDACTED]")
}
