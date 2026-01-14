// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeJSONSchema_DoesNotMutateInput(t *testing.T) {
	rawSchema := map[string]interface{}{
		"properties": map[string]interface{}{
			"foo": map[string]interface{}{
				"type": "string",
			},
		},
	}

    // Check initial state
    _, hasType := rawSchema["type"]
    assert.False(t, hasType)

	_, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	// Check if input was mutated
	_, hasTypeAfter := rawSchema["type"]
	assert.False(t, hasTypeAfter, "Input schema should not be mutated")
}

func TestSanitizeJSONSchema_DoesNotMutateNestedInput(t *testing.T) {
	rawSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"nested": map[string]interface{}{
				"properties": map[string]interface{}{
					"bar": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}

	_, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	// Check nested mutation
	props := rawSchema["properties"].(map[string]interface{})
	nested := props["nested"].(map[string]interface{})
	_, hasType := nested["type"]
	assert.False(t, hasType, "Nested input schema should not be mutated")
}

func TestSanitizeJSONSchema_DeepCopySlice(t *testing.T) {
	rawSchema := map[string]interface{}{
		"properties": map[string]interface{}{
			"foo": map[string]interface{}{
				"oneOf": []interface{}{
					map[string]interface{}{
						"type": "string",
					},
					map[string]interface{}{
						"type": "integer",
					},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	// Modify the raw schema slice to verify deep copy
	props := rawSchema["properties"].(map[string]interface{})
	foo := props["foo"].(map[string]interface{})
	oneOf := foo["oneOf"].([]interface{})
	oneOf0 := oneOf[0].(map[string]interface{})
	oneOf0["type"] = "modified"

	// Check sanitized schema
	sanitizedMap := sanitized.AsMap()
	sProps := sanitizedMap["properties"].(map[string]interface{})
	sFoo := sProps["foo"].(map[string]interface{})
	sOneOf := sFoo["oneOf"].([]interface{})
	sOneOf0 := sOneOf[0].(map[string]interface{})

	assert.Equal(t, "string", sOneOf0["type"], "Sanitized schema should not be affected by input mutation")
}

func TestSanitizeJSONSchema_DeepCopyNestedSlice(t *testing.T) {
	rawSchema := map[string]interface{}{
		"enum": []interface{}{
			[]interface{}{"a", "b"},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	// Modify raw
	enum := rawSchema["enum"].([]interface{})
	inner := enum[0].([]interface{})
	inner[0] = "modified"

	// Check sanitized
	sanitizedMap := sanitized.AsMap()
	sEnum := sanitizedMap["enum"].([]interface{})
	sInner := sEnum[0].([]interface{})
	assert.Equal(t, "a", sInner[0])
}
