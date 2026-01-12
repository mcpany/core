// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSensitiveKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		key      string
		expected bool
	}{
		{"api_key", true},
		{"API_KEY", true},
		{"access_token", true},
		{"password", true},
		{"client_secret", true},
		{"my_secret_value", true},
		{"auth_token", true},
		{"credential", true},
		{"private_key", true},
		{"username", false},
		{"email", false},
		{"url", false},
		{"description", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsSensitiveKey(tt.key))
		})
	}
}

func TestRedactMap(t *testing.T) {
	t.Parallel()
	input := map[string]interface{}{
		"username": "user1",
		"password": "secretpassword",
		"nested": map[string]interface{}{
			"api_key": "12345",
			"public":  "visible",
		},
		"list": []interface{}{
			map[string]interface{}{
				"token": "abcdef",
			},
			"normal_string",
		},
		"nested_slice": []interface{}{
			[]interface{}{
				map[string]interface{}{
					"secret": "hidden",
				},
			},
		},
	}

	redacted := RedactMap(input)

	assert.Equal(t, "user1", redacted["username"])
	assert.Equal(t, "[REDACTED]", redacted["password"])

	nested, ok := redacted["nested"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "[REDACTED]", nested["api_key"])
	assert.Equal(t, "visible", nested["public"])

	list, ok := redacted["list"].([]interface{})
	assert.True(t, ok)
	item0, ok := list[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "[REDACTED]", item0["token"])
	assert.Equal(t, "normal_string", list[1])
}

func TestRedactJSON(t *testing.T) {
	t.Parallel()
	t.Run("valid json object", func(t *testing.T) {
		input := `{"username": "user1", "password": "secretpassword"}`
		output := RedactJSON([]byte(input))

		var m map[string]interface{}
		err := json.Unmarshal(output, &m)
		assert.NoError(t, err)
		assert.Equal(t, "user1", m["username"])
		assert.Equal(t, "[REDACTED]", m["password"])
	})

	t.Run("valid json array", func(t *testing.T) {
		input := `[{"password": "secretpassword"}, {"public": "value"}]`
		output := RedactJSON([]byte(input))

		var s []interface{}
		err := json.Unmarshal(output, &s)
		assert.NoError(t, err)
		item0 := s[0].(map[string]interface{})
		assert.Equal(t, "[REDACTED]", item0["password"])
		item1 := s[1].(map[string]interface{})
		assert.Equal(t, "value", item1["public"])
	})

	t.Run("invalid json", func(t *testing.T) {
		input := `not valid json`
		output := RedactJSON([]byte(input))
		assert.Equal(t, []byte(input), output)
	})

	t.Run("large number precision", func(t *testing.T) {
		// A large integer that loses precision when converted to float64
		// 1234567890123456789 is large enough.
		input := `{"id": 1234567890123456789, "api_key": "secret"}`

		// We expect "api_key" to be redacted, but "id" to remain unchanged.
		// Note: we can't rely on key order so we parse it back
		output := RedactJSON([]byte(input))

		assert.Contains(t, string(output), "1234567890123456789")
		assert.Contains(t, string(output), "[REDACTED]")
	})

	t.Run("deeply nested json raw", func(t *testing.T) {
		// Test redactMapRaw recursive behavior with RawMessage
		// Note: redactMapRaw only recurses if value looks like object/array.

		// recurses into actual JSON objects/arrays.
		input2 := `{"nested": {"api_key": "secret", "deep": {"password": "pwd"}}}`
		output := RedactJSON([]byte(input2))

		var m map[string]interface{}
		err := json.Unmarshal(output, &m)
		assert.NoError(t, err)

		nested := m["nested"].(map[string]interface{})
		assert.Equal(t, "[REDACTED]", nested["api_key"])
		deep := nested["deep"].(map[string]interface{})
		assert.Equal(t, "[REDACTED]", deep["password"])
	})

	t.Run("array in object", func(t *testing.T) {
		input := `{"list": [{"password": "pwd"}]}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
	})

	t.Run("object in array", func(t *testing.T) {
		input := `[{"nested": {"password": "pwd"}}]`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
	})

	t.Run("key with escaped quotes", func(t *testing.T) {
		input := `{"api_key \"foo\"": "secret"}`
		output := RedactJSON([]byte(input))
		var m map[string]interface{}
		err := json.Unmarshal(output, &m)
		assert.NoError(t, err)

		val, ok := m["api_key \"foo\""]
		assert.True(t, ok, "key should exist")
		assert.Equal(t, "[REDACTED]", val)
	})

	t.Run("large allocation cap", func(t *testing.T) {
		// Create a large JSON input > 164KB to trigger the allocation cap logic.
		// 200KB string
		largeValue := strings.Repeat("a", 200*1024)
		// We need a sensitive key to trigger the allocation path
		input := `{"data": "` + largeValue + `", "api_key": "secret"}`

		output := RedactJSON([]byte(input))

		// Just ensure it processed correctly
		assert.Contains(t, string(output), `[REDACTED]`)
		assert.True(t, len(output) > 200*1024)
	})
}

func TestRedactJSON_FalsePositives(t *testing.T) {
	t.Parallel()
	// These tests verify that innocent words sharing a prefix with sensitive keys
	// are NOT redacted, unless they appear to be part of a composite key (CamelCase/PascalCase).

	tests := []struct {
		name     string
		input    string
		shouldRedact bool
	}{
		// False positives that should NOT be redacted
		{"author", `{"author": "John Doe"}`, false},        // "auth" + "or" (lowercase continuation)
		{"authority", `{"authority": "admin"}`, false},     // "auth" + "ority"
		{"authentic", `{"authentic": "true"}`, false},      // "auth" + "entic"
		{"tokens", `{"tokens": ["abc"]}`, false},           // "token" + "s" (lowercase continuation)
        {"AUTHORITY", `{"AUTHORITY": "public_info"}`, false}, // "AUTH" + "ORITY" (uppercase continuation)

		// True positives (Substring matching logic)
		{"CamelCase", `{"authToken": "123"}`, true},        // "auth" + "Token" (Uppercase boundary)
		{"PascalCase", `{"AuthToken": "123"}`, true},       // "Auth" + "Token"
		{"snake_case", `{"auth_token": "123"}`, true},      // "auth" + "_" (non-letter boundary)
		{"end of string", `{"user_auth": "123"}`, true},    // "auth" at end
		{"API_KEY", `{"API_KEY": "123"}`, true},            // exact upper
		{"api_key_val", `{"api_key_val": "123"}`, true},    // "api_key" + "_"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RedactJSON([]byte(tt.input))
			var m map[string]interface{}
			err := json.Unmarshal(output, &m)
			assert.NoError(t, err)

			// Find the single key in the map
			var val interface{}
			for _, v := range m {
				val = v
				break
			}

			if tt.shouldRedact {
				assert.Equal(t, "[REDACTED]", val, "Expected redaction for input: %s", tt.input)
			} else {
				assert.NotEqual(t, "[REDACTED]", val, "Expected NO redaction for input: %s", tt.input)
			}
		})
	}
}

func TestRedactJSON_Types(t *testing.T) {
	t.Parallel()
	t.Run("nested types", func(t *testing.T) {
		input := `{"api_key": {"foo": "bar"}, "token": ["a", "b"], "secret": 123, "secret_confidential": true, "password_hidden": null}`
		output := RedactJSON([]byte(input))

		var m map[string]interface{}
		err := json.Unmarshal(output, &m)
		require.NoError(t, err)

		assert.Equal(t, "[REDACTED]", m["api_key"])
		assert.Equal(t, "[REDACTED]", m["token"])
		assert.Equal(t, "[REDACTED]", m["secret"])
		assert.Equal(t, "[REDACTED]", m["secret_confidential"])
		assert.Equal(t, "[REDACTED]", m["password_hidden"])
	})

	t.Run("malformed json handled gracefully", func(t *testing.T) {
		// Unclosed string
		input := `{"api_key": "sec`
		output := RedactJSON([]byte(input))
		// Fail-safe behavior: redacts what it thinks is the value (until EOF) and appends placeholder
		expected := `{"api_key": "[REDACTED]"`
		assert.Equal(t, expected, string(output))
	})

	t.Run("escaped quotes in skipping", func(t *testing.T) {
		// String with escaped quotes
		input := `{"public": "some \"string\"", "api_key": "secret"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), `some \"string\"`)
		assert.Contains(t, string(output), `[REDACTED]`)
	})

	t.Run("nested object skipping", func(t *testing.T) {
		input := `{"public": {"nested": "value"}, "api_key": "secret"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), `{"nested": "value"}`)
		assert.Contains(t, string(output), `[REDACTED]`)
	})

	t.Run("nested array skipping", func(t *testing.T) {
		input := `{"public": ["a", "b"], "api_key": "secret"}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), `["a", "b"]`)
		assert.Contains(t, string(output), `[REDACTED]`)
	})
}

func TestScanForSensitiveKeys_JSON(t *testing.T) {
	t.Parallel()
	// Test the scanJSONForSensitiveKeys path (validateKeyContext=true)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"simple key", `{"api_key": "value"}`, true},
		{"nested key", `{"nested": {"api_key": "value"}}`, true},
		{"key in string value (should be ignored)", `{"public": "this contains api_key but is a value"}`, false},
		{"escaped quote in string", `{"public": "some \" quote", "api_key": "secret"}`, true},
		{"malformed json (ignored)", `not json`, false},
		{"key without colon", `{"api_key"}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanForSensitiveKeys([]byte(tt.input), true)
			assert.Equal(t, tt.expected, result)
		})
	}
}
