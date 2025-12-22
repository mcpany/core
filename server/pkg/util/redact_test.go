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
}
