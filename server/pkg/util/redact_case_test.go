package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_CaseSensitivityBug(t *testing.T) {
	// usage of uppercase key should be redacted.
	// value must not contain any sensitive key strings (like "secret")
	// to avoid triggering the optimization check by accident.
	input := `{"API_KEY": "safe_value", "public": "value"}`
	output := RedactJSON([]byte(input))

	var m map[string]interface{}
	err := json.Unmarshal(output, &m)
	assert.NoError(t, err)

	assert.Equal(t, "[REDACTED]", m["API_KEY"], "API_KEY should be redacted")
	assert.Equal(t, "value", m["public"])
}
