// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeJSONSchema_CircularReference_Safe(t *testing.T) {
	// Create a circular reference
	root := make(map[string]interface{})
	child := make(map[string]interface{})
	root["properties"] = child
	child["parent"] = root

	// This should NOT crash now, but return an error
	_, err := SanitizeJSONSchema(root)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum recursion depth")
}

func TestSanitizeJSONSchema_ItemsArray(t *testing.T) {
	// Schema with items as an array (tuple validation)
	input := map[string]interface{}{
		"type": "array",
		"items": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"foo": map[string]interface{}{
						"type": "string",
					},
				},
			},
			map[string]interface{}{
				"type": "string",
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)

	sMap := sanitized.AsMap()
	items := sMap["items"].([]interface{})
	require.Len(t, items, 2)

	// first item should have been sanitized (missing type: object added)
	item0 := items[0].(map[string]interface{})
	assert.Equal(t, "object", item0["type"])

	// second item should be untouched
	item1 := items[1].(map[string]interface{})
	assert.Equal(t, "string", item1["type"])
}

func TestSanitizeJSONSchema_AdditionalProperties(t *testing.T) {
	input := map[string]interface{}{
		"type": "object",
		"additionalProperties": map[string]interface{}{
			"properties": map[string]interface{}{
				"bar": map[string]interface{}{
					"type": "integer",
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)

	sMap := sanitized.AsMap()
	addProps := sMap["additionalProperties"].(map[string]interface{})

	// Should have added type: object
	assert.Equal(t, "object", addProps["type"])
}

func TestSanitizeJSONSchema_Combinators(t *testing.T) {
	input := map[string]interface{}{
		"oneOf": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"a": map[string]interface{}{"type": "string"},
				},
			},
			map[string]interface{}{
				"type": "string",
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)

	sMap := sanitized.AsMap()
	oneOf := sMap["oneOf"].([]interface{})
	require.Len(t, oneOf, 2)

	// First option should be sanitized
	opt0 := oneOf[0].(map[string]interface{})
	assert.Equal(t, "object", opt0["type"])
}

func TestSanitizeJSONSchema_Definitions(t *testing.T) {
	input := map[string]interface{}{
		"$defs": map[string]interface{}{
			"myType": map[string]interface{}{
				"properties": map[string]interface{}{
					"b": map[string]interface{}{"type": "boolean"},
				},
			},
		},
		"definitions": map[string]interface{}{
			"oldType": map[string]interface{}{
				"properties": map[string]interface{}{
					"c": map[string]interface{}{"type": "number"},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)

	sMap := sanitized.AsMap()

	defs := sMap["$defs"].(map[string]interface{})
	myType := defs["myType"].(map[string]interface{})
	assert.Equal(t, "object", myType["type"])

	oldDefs := sMap["definitions"].(map[string]interface{})
	oldType := oldDefs["oldType"].(map[string]interface{})
	assert.Equal(t, "object", oldType["type"])
}

func TestSanitizeJSONSchema_DeepNesting(t *testing.T) {
	// Build a deeply nested object
	root := make(map[string]interface{})
	current := root
	// Go deeper than maxRecursionDepth (100)
	for i := 0; i < 110; i++ {
		next := make(map[string]interface{})
		current["next"] = next
		current = next
	}

	_, err := SanitizeJSONSchema(root)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum recursion depth")
}
