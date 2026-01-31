// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeSchema_Definitions(t *testing.T) {
	// Schema with a definition that is missing "type": "object"
	rawSchema := map[string]interface{}{
		"definitions": map[string]interface{}{
			"foo": map[string]interface{}{
				"properties": map[string]interface{}{
					"bar": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	defs := sanitizedMap["definitions"].(map[string]interface{})
	foo := defs["foo"].(map[string]interface{})

	// This assertion should fail if definitions are ignored
	assert.Equal(t, "object", foo["type"], "Expected 'type': 'object' to be added to definition")
}

func TestSanitizeSchema_Defs(t *testing.T) {
	// Schema with a $defs that is missing "type": "object"
	rawSchema := map[string]interface{}{
		"$defs": map[string]interface{}{
			"foo": map[string]interface{}{
				"properties": map[string]interface{}{
					"bar": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	defs := sanitizedMap["$defs"].(map[string]interface{})
	foo := defs["foo"].(map[string]interface{})

	assert.Equal(t, "object", foo["type"], "Expected 'type': 'object' to be added to $defs")
}

func TestSanitizeSchema_OneOf(t *testing.T) {
	rawSchema := map[string]interface{}{
		"oneOf": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"foo": map[string]interface{}{"type": "string"},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	oneOf := sanitizedMap["oneOf"].([]interface{})
	first := oneOf[0].(map[string]interface{})

	assert.Equal(t, "object", first["type"], "Expected 'type': 'object' to be added to oneOf element")
}

func TestSanitizeSchema_AnyOf(t *testing.T) {
	rawSchema := map[string]interface{}{
		"anyOf": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"foo": map[string]interface{}{"type": "string"},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	anyOf := sanitizedMap["anyOf"].([]interface{})
	first := anyOf[0].(map[string]interface{})

	assert.Equal(t, "object", first["type"], "Expected 'type': 'object' to be added to anyOf element")
}

func TestSanitizeSchema_AllOf(t *testing.T) {
	rawSchema := map[string]interface{}{
		"allOf": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"foo": map[string]interface{}{"type": "string"},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	allOf := sanitizedMap["allOf"].([]interface{})
	first := allOf[0].(map[string]interface{})

	assert.Equal(t, "object", first["type"], "Expected 'type': 'object' to be added to allOf element")
}

func TestSanitizeSchema_Items_Array(t *testing.T) {
	// items can be a schema or an array of schemas (for tuple validation)
	rawSchema := map[string]interface{}{
		"type": "array",
		"items": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"foo": map[string]interface{}{"type": "string"},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	items := sanitizedMap["items"].([]interface{})
	first := items[0].(map[string]interface{})

	assert.Equal(t, "object", first["type"], "Expected 'type': 'object' to be added to items array element")
}

func TestSanitizeSchema_AdditionalProperties(t *testing.T) {
	// additionalProperties can be a schema
	rawSchema := map[string]interface{}{
		"type": "object",
		"additionalProperties": map[string]interface{}{
			"properties": map[string]interface{}{
				"foo": map[string]interface{}{"type": "string"},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	addProps := sanitizedMap["additionalProperties"].(map[string]interface{})

	assert.Equal(t, "object", addProps["type"], "Expected 'type': 'object' to be added to additionalProperties schema")
}

func TestSanitizeSchema_DeeplyNested(t *testing.T) {
	// deeply nested mix
	rawSchema := map[string]interface{}{
		"definitions": map[string]interface{}{
			"myType": map[string]interface{}{
				"oneOf": []interface{}{
					map[string]interface{}{
						"properties": map[string]interface{}{
							"nested": map[string]interface{}{"type": "string"},
						},
					},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	defs := sanitizedMap["definitions"].(map[string]interface{})
	myType := defs["myType"].(map[string]interface{})
	oneOf := myType["oneOf"].([]interface{})
	first := oneOf[0].(map[string]interface{})

	assert.Equal(t, "object", first["type"], "Expected 'type': 'object' to be added to deeply nested schema")
}
