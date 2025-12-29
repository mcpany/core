// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSensitiveKey(t *testing.T) {
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
