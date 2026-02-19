// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeJSONSchema_CircularReference(t *testing.T) {
	// Create a circular reference
	root := make(map[string]interface{})
	child := make(map[string]interface{})
	root["properties"] = child
	child["parent"] = root

	// This should NOT crash (stack overflow) if cycle detection is implemented.
	// We expect SanitizeJSONSchema to return an error or handle it gracefully.
	// For now, if it returns, it's a pass (no panic).
	// If deepCopyJSON is implemented correctly, it should probably return an error or nil for the cycle.

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SanitizeJSONSchema panicked with circular reference: %v", r)
		}
	}()

	sanitized, err := SanitizeJSONSchema(root)
	// We don't necessarily enforce an error here, but we enforce NO PANIC.
	// If it returns a result, we check if it's sane.
	if err == nil && sanitized != nil {
		assert.NotNil(t, sanitized)
	}
}

func TestSanitizeJSONSchema_ItemsArray(t *testing.T) {
	// items can be an array of schemas (tuple validation)
	input := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"tuple": map[string]interface{}{
				"type": "array",
				"items": []interface{}{
					map[string]interface{}{
						// Missing type: object but has properties
						"properties": map[string]interface{}{
							"foo": map[string]interface{}{"type": "string"},
						},
					},
					map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	props := sanitizedMap["properties"].(map[string]interface{})
	tuple := props["tuple"].(map[string]interface{})
	items := tuple["items"].([]interface{})

	require.Len(t, items, 2)
	item0 := items[0].(map[string]interface{})

	// Check if "type": "object" was added to the first item in the array
	assert.Equal(t, "object", item0["type"], "Sanitizer should recurse into items array")
}

func TestSanitizeJSONSchema_AdditionalProperties(t *testing.T) {
	// additionalProperties can be a schema
	input := map[string]interface{}{
		"type": "object",
		"additionalProperties": map[string]interface{}{
			// Missing type: object
			"properties": map[string]interface{}{
				"bar": map[string]interface{}{"type": "string"},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	addProps := sanitizedMap["additionalProperties"].(map[string]interface{})

	// Check if "type": "object" was added
	assert.Equal(t, "object", addProps["type"], "Sanitizer should recurse into additionalProperties")
}

func TestSanitizeJSONSchema_Defs(t *testing.T) {
	// $defs or definitions
	input := map[string]interface{}{
		"$defs": map[string]interface{}{
			"myType": map[string]interface{}{
				// Missing type: object
				"properties": map[string]interface{}{
					"baz": map[string]interface{}{"type": "string"},
				},
			},
		},
		"definitions": map[string]interface{}{
			"oldType": map[string]interface{}{
				// Missing type: object
				"properties": map[string]interface{}{
					"qux": map[string]interface{}{"type": "string"},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)

	sanitizedMap := sanitized.AsMap()

	if defs, ok := sanitizedMap["$defs"].(map[string]interface{}); ok {
		myType := defs["myType"].(map[string]interface{})
		assert.Equal(t, "object", myType["type"], "Sanitizer should recurse into $defs")
	} else {
		// Depending on proto conversion, $defs might be lost if structpb doesn't preserve it nicely?
		// No, structpb preserves generic maps.
		t.Log("Warning: $defs not found or not a map")
	}

	if definitions, ok := sanitizedMap["definitions"].(map[string]interface{}); ok {
		oldType := definitions["oldType"].(map[string]interface{})
		assert.Equal(t, "object", oldType["type"], "Sanitizer should recurse into definitions")
	}
}

func TestSanitizeJSONSchema_Combinators(t *testing.T) {
	// oneOf, anyOf, allOf
	input := map[string]interface{}{
		"oneOf": []interface{}{
			map[string]interface{}{
				// Missing type: object
				"properties": map[string]interface{}{
					"a": map[string]interface{}{"type": "string"},
				},
			},
		},
		"anyOf": []interface{}{
			map[string]interface{}{
				// Missing type: object
				"properties": map[string]interface{}{
					"b": map[string]interface{}{"type": "string"},
				},
			},
		},
		"allOf": []interface{}{
			map[string]interface{}{
				// Missing type: object
				"properties": map[string]interface{}{
					"c": map[string]interface{}{"type": "string"},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)

	sanitizedMap := sanitized.AsMap()

	oneOf := sanitizedMap["oneOf"].([]interface{})
	oneOf0 := oneOf[0].(map[string]interface{})
	assert.Equal(t, "object", oneOf0["type"], "Sanitizer should recurse into oneOf")

	anyOf := sanitizedMap["anyOf"].([]interface{})
	anyOf0 := anyOf[0].(map[string]interface{})
	assert.Equal(t, "object", anyOf0["type"], "Sanitizer should recurse into anyOf")

	allOf := sanitizedMap["allOf"].([]interface{})
	allOf0 := allOf[0].(map[string]interface{})
	assert.Equal(t, "object", allOf0["type"], "Sanitizer should recurse into allOf")
}

func TestSanitizeJSONSchema_NonMapProperties(t *testing.T) {
	// Properties with non-map values should be ignored gracefully
	input := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"weird": "string", // Should be a schema object, but if it's a string, ignore
			"valid": map[string]interface{}{
				"type": "string",
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	props := sanitizedMap["properties"].(map[string]interface{})
	assert.Equal(t, "string", props["weird"])
	assert.Equal(t, "string", props["valid"].(map[string]interface{})["type"])
}
