package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_CaseSensitivityBug(t *testing.T) {
	// usage of uppercase key should be redacted.
	// The value should not contain any sensitive keywords to avoid false positives in the optimization check.
	input := `{"API_KEY": "safe_value", "public": "value"}`
	output := RedactJSON([]byte(input))

	var m map[string]interface{}
	err := json.Unmarshal(output, &m)
	assert.NoError(t, err)

	// If the bug exists, API_KEY will not be redacted because
	// RedactJSON optimization checks for "api_key" (lowercase) in bytes,
	// and finding none (and no other sensitive keywords in values), returns input as is.
	assert.Equal(t, "[REDACTED]", m["API_KEY"], "API_KEY should be redacted")
	assert.Equal(t, "value", m["public"])
}
