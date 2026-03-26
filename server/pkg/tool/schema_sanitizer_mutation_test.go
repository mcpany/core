// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSanitizeJSONSchema_DoesNotMutateInput ...
// Summary: TestSanitizeJSONSchema_DoesNotMutateInput
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestSanitizeJSONSchema_DoesNotMutateNestedInput ...
// Summary: TestSanitizeJSONSchema_DoesNotMutateNestedInput
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
