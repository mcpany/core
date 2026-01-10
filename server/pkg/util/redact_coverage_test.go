package util //nolint:revive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_Coverage(t *testing.T) {
	// Case 14: Value is an array
	t.Run("array value", func(t *testing.T) {
		input := `{"password": ["val1", "val2"]}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
		assert.NotContains(t, string(output), "val1")
	})

	// Case 15: Value is boolean true
	t.Run("boolean true", func(t *testing.T) {
		input := `{"password": true}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
	})

	// Case 16: Value is boolean false
	t.Run("boolean false", func(t *testing.T) {
		input := `{"password": false}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
	})

	// Case 17: Value is null
	t.Run("null value", func(t *testing.T) {
		input := `{"password": null}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
	})

	// Case 18: Array with nested structure
	t.Run("array nested", func(t *testing.T) {
		input := `{"password": [{"k":"v"}]}`
		output := RedactJSON([]byte(input))
		assert.Contains(t, string(output), "[REDACTED]")
		assert.NotContains(t, string(output), "k")
	})
}
