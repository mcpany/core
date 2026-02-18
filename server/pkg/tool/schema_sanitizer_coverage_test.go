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

	// This should not panic
	sanitized, err := SanitizeJSONSchema(root)

	// We expect either nil (cycle broken) or a truncated structure.
	// Since the cycle is immediate, it might just return nil for the cycle link.
	// The important thing is NO PANIC.
	require.NoError(t, err)
	assert.NotNil(t, sanitized)

	m := sanitized.AsMap()
	assert.Contains(t, m, "properties")
	props := m["properties"].(map[string]interface{})

	// The cycle link "parent" should be gone or null
	// Wait, deepCopyJSON returns nil for cycle. So "parent": nil.
	// sanitizeJSONSchemaInPlace also returns nil for cycle.
	// structpb handles nil values by omitting them or using null value?
	// structpb.NewStruct(map) handles nil values as NullValue.

	parent, ok := props["parent"]
	if ok {
		assert.Nil(t, parent, "Cycle link should be nil")
	}
}

func TestSanitizeJSONSchema_ItemsArray(t *testing.T) {
	input := map[string]interface{}{
		"type": "array",
		"items": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"foo": map[string]interface{}{"type": "string"},
				},
			},
			map[string]interface{}{
				"type": "string",
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)

	m := sanitized.AsMap()
	items := m["items"].([]interface{})
	require.Len(t, items, 2)

	item0 := items[0].(map[string]interface{})
	assert.Equal(t, "object", item0["type"], "Should add type: object to item 0")

	item1 := items[1].(map[string]interface{})
	assert.Equal(t, "string", item1["type"])
}

func TestSanitizeJSONSchema_AdditionalProperties(t *testing.T) {
	input := map[string]interface{}{
		"type": "object",
		"additionalProperties": map[string]interface{}{
			"properties": map[string]interface{}{
				"bar": map[string]interface{}{"type": "integer"},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)

	m := sanitized.AsMap()
	addProps := m["additionalProperties"].(map[string]interface{})
	assert.Equal(t, "object", addProps["type"], "Should add type: object to additionalProperties schema")
}

func TestSanitizeJSONSchema_AllOf_AnyOf_OneOf(t *testing.T) {
	input := map[string]interface{}{
		"allOf": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{"a": map[string]interface{}{"type": "string"}},
			},
		},
		"anyOf": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{"b": map[string]interface{}{"type": "string"}},
			},
		},
		"oneOf": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{"c": map[string]interface{}{"type": "string"}},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)
	m := sanitized.AsMap()

	allOf := m["allOf"].([]interface{})
	assert.Equal(t, "object", allOf[0].(map[string]interface{})["type"])

	anyOf := m["anyOf"].([]interface{})
	assert.Equal(t, "object", anyOf[0].(map[string]interface{})["type"])

	oneOf := m["oneOf"].([]interface{})
	assert.Equal(t, "object", oneOf[0].(map[string]interface{})["type"])
}

func TestSanitizeJSONSchema_Definitions(t *testing.T) {
	input := map[string]interface{}{
		"$defs": map[string]interface{}{
			"def1": map[string]interface{}{
				"properties": map[string]interface{}{"x": map[string]interface{}{"type": "string"}},
			},
		},
		"definitions": map[string]interface{}{
			"def2": map[string]interface{}{
				"properties": map[string]interface{}{"y": map[string]interface{}{"type": "string"}},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)
	m := sanitized.AsMap()

	defs := m["$defs"].(map[string]interface{})
	assert.Equal(t, "object", defs["def1"].(map[string]interface{})["type"])

	definitions := m["definitions"].(map[string]interface{})
	assert.Equal(t, "object", definitions["def2"].(map[string]interface{})["type"])
}

func TestSanitizeJSONSchema_DeeplyNested(t *testing.T) {
	// Generate a deep structure
	current := map[string]interface{}{
		"type": "string",
	}
	for i := 0; i < 50; i++ {
		current = map[string]interface{}{
			"properties": map[string]interface{}{
				"next": current,
			},
		}
	}

	sanitized, err := SanitizeJSONSchema(current)
	require.NoError(t, err)
	assert.NotNil(t, sanitized)

	// Verify deep nesting is preserved (at least mostly)
	m := sanitized.AsMap()
	for i := 0; i < 50; i++ {
		assert.Equal(t, "object", m["type"], "Level %d should have type object", i)
		props := m["properties"].(map[string]interface{})
		m = props["next"].(map[string]interface{})
	}
	assert.Equal(t, "string", m["type"])
}
