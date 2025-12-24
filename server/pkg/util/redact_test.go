// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"encoding/json"
	"strings"
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
}

func TestRedactJSON_SmartDetection(t *testing.T) {
	// Case 1: False positive - sensitive key text in VALUE.
	// Should NOT redact (and more importantly, should skip parsing in our optimization)
	// But we can verify output is identical.
	inputFalsePositive := `{"description": "this contains api_key but not as a key"}`
	outputFP := RedactJSON([]byte(inputFalsePositive))
	if !bytes.Equal(outputFP, []byte(inputFalsePositive)) {
		t.Errorf("False positive case modified input: got %s, want %s", outputFP, inputFalsePositive)
	}

	// Case 2: True positive - sensitive key IS a key.
	// Should redact.
	inputTruePositive := `{"api_key": "secret"}`
	expectedTP := `"[REDACTED]"`
	outputTP := RedactJSON([]byte(inputTruePositive))
	if !strings.Contains(string(outputTP), expectedTP) {
		t.Errorf("True positive case failed: got %s, want to contain %s", outputTP, expectedTP)
	}

	// Case 3: True positive - sensitive key IS a key (with spaces).
	inputTPSpaces := `{"api_key" : "secret"}`
	expectedTPSpaces := `"[REDACTED]"`
	outputTPSpaces := RedactJSON([]byte(inputTPSpaces))
	if !strings.Contains(string(outputTPSpaces), expectedTPSpaces) {
		t.Errorf("True positive with spaces case failed: got %s, want to contain %s", outputTPSpaces, expectedTPSpaces)
	}

	// Case 4: True positive - nested key.
	inputNested := `{"config": {"token": "secret"}}`
	expectedNested := `"[REDACTED]"`
	outputNested := RedactJSON([]byte(inputNested))
	if !strings.Contains(string(outputNested), expectedNested) {
		t.Errorf("Nested case failed: got %s, want to contain %s", outputNested, expectedNested)
	}

	// Case 5: Edge case - escaped quotes in value.
	// "description": "some \"api_key\" in quotes"
	// This is a VALUE. Should NOT redact.
	inputEscaped := `{"description": "some \"api_key\" in quotes"}`
	outputEscaped := RedactJSON([]byte(inputEscaped))
	if !bytes.Equal(outputEscaped, []byte(inputEscaped)) {
		t.Errorf("Escaped quote case modified input: got %s, want %s", outputEscaped, inputEscaped)
	}

	// Case 6: Edge case - "key": "value", "api_key": "secret"
	// Ensure we don't stop at first non-match.
	inputMultiple := `{"name": "test", "api_key": "secret"}`
	outputMultiple := RedactJSON([]byte(inputMultiple))
	if !strings.Contains(string(outputMultiple), `"[REDACTED]"`) {
		t.Errorf("Multiple keys case failed to redact: got %s", outputMultiple)
	}
	if !strings.Contains(string(outputMultiple), `"name":"test"`) {
		t.Errorf("Multiple keys case missing other fields: got %s", outputMultiple)
	}
}
