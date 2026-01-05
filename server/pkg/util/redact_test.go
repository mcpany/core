// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
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
