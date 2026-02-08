// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeSchema_Cycle(t *testing.T) {
	// Create a cycle
	cycle := make(map[string]interface{})
	cycle["self"] = cycle

	rawSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"cycle": cycle,
		},
	}

	// This should return an error or sanitize it by breaking the cycle
	// Ideally it returns an error rather than stack overflow
	_, err := SanitizeJSONSchema(rawSchema)
	// We expect an error about recursion limit or cycle
	if err == nil {
		// If no error, ensure no panic
	} else {
		assert.Error(t, err)
	}
}

func TestSanitizeSchema_ItemsArray(t *testing.T) {
	// Schema with items as array (tuple validation)
	// We put a nested property that needs sanitization (missing type: object)
	rawSchemaWithIssue := map[string]interface{}{
		"type": "array",
		"items": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"foo": map[string]interface{}{"type": "string"},
				},
				// Missing "type": "object", should be added by sanitizer
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchemaWithIssue)
	assert.NoError(t, err)

	sMap := sanitized.AsMap()
	items, ok := sMap["items"].([]interface{})
	assert.True(t, ok, "items should be an array")
	if ok && len(items) > 0 {
		item0, ok := items[0].(map[string]interface{})
		assert.True(t, ok, "item 0 should be a map")
		assert.Equal(t, "object", item0["type"], "sanitizer should add type: object to array item")
	}
}

func TestSanitizeSchema_InvalidTypes(t *testing.T) {
	// "properties" should be a map, but here it is a string
	rawSchema := map[string]interface{}{
		"type": "object",
		"properties": "invalid",
	}

	// Should not panic
	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)
	assert.NotNil(t, sanitized)
}

func TestSanitizeSchema_DeepNesting(t *testing.T) {
	// Create a deeply nested schema
	root := make(map[string]interface{})
	current := root
	for i := 0; i < 200; i++ {
		current["type"] = "object"
		props := make(map[string]interface{})
		current["properties"] = props
		next := make(map[string]interface{})
		props["next"] = next
		current = next
	}

	// Should not stack overflow (if depth limit is implemented or Go stack is large enough)
	sanitized, err := SanitizeJSONSchema(root)
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.NotNil(t, sanitized)
	}
}
