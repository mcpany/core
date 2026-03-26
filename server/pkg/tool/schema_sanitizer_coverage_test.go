// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSanitizeJSONSchema_CircularReference_Safe ...
// Summary: TestSanitizeJSONSchema_CircularReference_Safe
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestSanitizeJSONSchema_ItemsArray ...
// Summary: TestSanitizeJSONSchema_ItemsArray
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestSanitizeJSONSchema_AdditionalProperties ...
// Summary: TestSanitizeJSONSchema_AdditionalProperties
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestSanitizeJSONSchema_Combinators ...
// Summary: TestSanitizeJSONSchema_Combinators
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestSanitizeJSONSchema_Definitions ...
// Summary: TestSanitizeJSONSchema_Definitions
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestSanitizeJSONSchema_DeepNesting ...
// Summary: TestSanitizeJSONSchema_DeepNesting
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
