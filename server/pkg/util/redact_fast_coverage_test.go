package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRedactFast_Coverage(t *testing.T) {
	// Cover skipLiteral
	// true, false, null
	assert.Equal(t, []byte(`{"a": true, "b": false, "c": null}`), redactJSONFast([]byte(`{"a": true, "b": false, "c": null}`)))

	// Cover skipNumber
	assert.Equal(t, []byte(`{"a": 123}`), redactJSONFast([]byte(`{"a": 123}`)))
	assert.Equal(t, []byte(`{"a": 123.456}`), redactJSONFast([]byte(`{"a": 123.456}`)))

	// Cover skipObject nested with strings
	// Ensure we don't break on "}" inside string
	input := `{"a": {"b": "}"}}`
	assert.Equal(t, []byte(input), redactJSONFast([]byte(input)))

	// Cover skipArray
	inputArr := `{"a": ["b", "c"]}`
	assert.Equal(t, []byte(inputArr), redactJSONFast([]byte(inputArr)))

	// Cover sensitive key with array value
	// {"api_key": ["a", "b"]} -> {"api_key": "[REDACTED]"}
	inputSens := `{"api_key": ["a", "b"]}`
	expectedSens := `{"api_key": "[REDACTED]"}`
	assert.Equal(t, []byte(expectedSens), redactJSONFast([]byte(inputSens)))

	// Cover sensitive key with object value
	// {"api_key": {"a": "b"}} -> {"api_key": "[REDACTED]"}
	inputSensObj := `{"api_key": {"a": "b"}}`
	expectedSensObj := `{"api_key": "[REDACTED]"}`
	assert.Equal(t, []byte(expectedSensObj), redactJSONFast([]byte(inputSensObj)))
}
