// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeSchema_MissingTypeObject(t *testing.T) {
	// Represents a schema that is missing "type": "object" but has "properties"
	rawSchema := map[string]interface{}{
		"properties": map[string]interface{}{
			"foo": map[string]interface{}{
				"type": "string",
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	assert.Equal(t, "object", sanitizedMap["type"])
}

func TestSanitizeSchema_Recursive(t *testing.T) {
	// Represents a schema that is missing "type": "object" but has "properties" nested
	rawSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
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
	props := sanitizedMap["properties"].(map[string]interface{})
	foo := props["foo"].(map[string]interface{})
	assert.Equal(t, "object", foo["type"])
}
