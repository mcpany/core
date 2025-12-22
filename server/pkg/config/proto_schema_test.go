package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileSchema_Error(t *testing.T) {
	// Function values cannot be marshaled to JSON
	badMap := map[string]interface{}{
		"bad": func() {},
	}

	schema, err := compileSchema(badMap)
	assert.Error(t, err)
	assert.Nil(t, schema)
	assert.Contains(t, err.Error(), "json: unsupported type")
}

func TestCompileSchema_InvalidSchema(t *testing.T) {
	// Valid JSON but invalid schema ($id must be a string)
	// This triggers AddResource parsing failure
	badSchema := map[string]interface{}{
		"$id": 123,
	}

	schema, err := compileSchema(badSchema)
	assert.Error(t, err)
	assert.Nil(t, schema)
	// This usually returns "json: cannot unmarshal number into Go struct field Schema.id of type string" or similar
}
