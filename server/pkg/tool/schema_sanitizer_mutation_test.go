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
