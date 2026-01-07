package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_QuoteInKey_Regression(t *testing.T) {
	t.Parallel()
	// "password" is a sensitive key.
	// We use a key "password\"" (password followed by quote).
	// Since "password" is a prefix, it should match.
	// The isKey check in scanForSensitiveKeys would previously fail because it would find a quote but no colon.
	// With the fix, we skip isKey check when scanning keys directly.

	input := `{"password\"": "secret"}`
	output := RedactJSON([]byte(input))

	var m map[string]interface{}
	err := json.Unmarshal(output, &m)
	assert.NoError(t, err)

	// We expect "password\"" to be redacted because it starts with "password".
	// The value "secret" should be [REDACTED].
	assert.Equal(t, "[REDACTED]", m["password\""])
}
