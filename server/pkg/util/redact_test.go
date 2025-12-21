// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON(t *testing.T) {
	t.Run("redact_object", func(t *testing.T) {
		input := `{"name": "test", "api_key": "secret", "password": "123", "nested": {"token": "abc"}}`
		expected := `{"api_key":"[REDACTED]","name":"test","nested":{"token":"[REDACTED]"},"password":"[REDACTED]"}`

		output := RedactJSON([]byte(input))

		// Unmarshal to map to compare because JSON key order is undefined
		var outMap map[string]interface{}
		var expMap map[string]interface{}

		err := json.Unmarshal(output, &outMap)
		assert.NoError(t, err)

		err = json.Unmarshal([]byte(expected), &expMap)
		assert.NoError(t, err)

		assert.Equal(t, expMap, outMap)
	})

	t.Run("redact_array", func(t *testing.T) {
		input := `[{"api_key": "secret"}, {"safe": "value"}]`
		expected := `[{"api_key":"[REDACTED]"},{"safe":"value"}]`

		output := RedactJSON([]byte(input))

		var outSlice []interface{}
		var expSlice []interface{}

		err := json.Unmarshal(output, &outSlice)
		assert.NoError(t, err)

		err = json.Unmarshal([]byte(expected), &expSlice)
		assert.NoError(t, err)

		assert.Equal(t, expSlice, outSlice)
	})

	t.Run("invalid_json", func(t *testing.T) {
		input := `invalid json`
		output := RedactJSON([]byte(input))
		assert.Equal(t, []byte(input), output)
	})
}

func TestRedactMap(t *testing.T) {
	input := map[string]interface{}{
		"api_key": "secret",
		"nested": map[string]interface{}{
			"token": "abc",
			"safe":  "value",
		},
		"list": []interface{}{
			map[string]interface{}{"password": "123"},
		},
	}

	expected := map[string]interface{}{
		"api_key": "[REDACTED]",
		"nested": map[string]interface{}{
			"token": "[REDACTED]",
			"safe":  "value",
		},
		"list": []interface{}{
			map[string]interface{}{"password": "[REDACTED]"},
		},
	}

	output := RedactMap(input)
	assert.Equal(t, expected, output)
}

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"api_key", true},
		{"API_KEY", true},
		{"my_token", true},
		{"password", true},
		{"secret_value", true},
		{"auth_header", true},
		{"name", false},
		{"description", false},
		{"url", false},
	}

	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsSensitiveKey(tc.key))
		})
	}
}
