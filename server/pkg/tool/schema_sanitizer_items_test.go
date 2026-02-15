package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeJSONSchema_RecursiveItems(t *testing.T) {
	// A schema with an array whose items are objects with properties, but missing "type": "object"
	input := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"list": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					// Missing "type": "object"
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(input)
	require.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	props := sanitizedMap["properties"].(map[string]interface{})
	list := props["list"].(map[string]interface{})
	items := list["items"].(map[string]interface{})

	// Check if "type": "object" was added to items
	assert.Equal(t, "object", items["type"], "Sanitizer should add 'type': 'object' to items schema if 'properties' is present")
}
