package mcpserver

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLazyRedact_LogValue_Comments(t *testing.T) {
	// This test simulates the logging of request arguments that contain comments with braces.
	// This was causing invalid redaction/JSON corruption.

	input := []byte(`{
		"secret": {
			"a": 1
			// comment with } brace
		},
		"public": "visible"
	}`)

	lr := LazyRedact(input)
	val := lr.LogValue()

	assert.Equal(t, slog.KindString, val.Kind())

	// slog.Value.String() returns the string value if Kind is String.
	// Note: For other kinds, it might return a representation.
	// But LazyRedact returns slog.StringValue(...).

	redactedStr := val.String()

    // We verify that the output contains the public part, which proves that
    // the redaction didn't stop early or corrupt the rest of the JSON.
	assert.Contains(t, redactedStr, `"secret": "[REDACTED]"`)
	assert.Contains(t, redactedStr, `"public": "visible"`)
}
