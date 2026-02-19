// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeJSONSchema_Coverage(t *testing.T) {
	t.Parallel()

	t.Run("Circular Reference", func(t *testing.T) {
		// This test is expected to fail (stack overflow) until the fix is implemented.
		// We use a goroutine and recover to catch panics if possible, but stack overflow
		// usually crashes the runtime hard.
		// For the purpose of TDD, we can skip it or just let it crash locally.
		// However, to keep the test suite runnable, I will comment it out or skip it
		// until I implement the fix. But the plan says "Create Comprehensive Test Suite"
		// then "Implement Fixes". If I add a crashing test now, I can't run other tests.
		// So I will add it but skip it, then unskip in the fix step?
		// Better: I will implement the test but disable it with a flag or just leave it
		// commented out with a TODO.
		// actually, I can just not run it until the fix step.

		root := make(map[string]interface{})
		child := make(map[string]interface{})
		root["properties"] = child
		child["parent"] = root

		// This should now run without crashing (pruning at depth 500)
		sanitized, err := SanitizeJSONSchema(root)
		assert.NoError(t, err)
		assert.NotNil(t, sanitized)
	})

	t.Run("Items Array", func(t *testing.T) {
		// Schema with "items" as an array of schemas (tuple validation)
		input := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"tuple": map[string]interface{}{
					"type": "array",
					"items": []interface{}{
						map[string]interface{}{
							// Missing type: object
							"properties": map[string]interface{}{
								"a": map[string]interface{}{"type": "string"},
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

		// Verify that the first item in the tuple got sanitized (type: object added)
		sMap := sanitized.AsMap()
		tuple := sMap["properties"].(map[string]interface{})["tuple"].(map[string]interface{})
		items := tuple["items"].([]interface{})
		item0 := items[0].(map[string]interface{})

		// Current implementation fails this assertion because it ignores items array
		assert.Equal(t, "object", item0["type"])
	})

	t.Run("AdditionalProperties Schema", func(t *testing.T) {
		input := map[string]interface{}{
			"type": "object",
			"additionalProperties": map[string]interface{}{
				// Missing type: object
				"properties": map[string]interface{}{
					"foo": map[string]interface{}{"type": "string"},
				},
			},
		}

		sanitized, err := SanitizeJSONSchema(input)
		require.NoError(t, err)

		sMap := sanitized.AsMap()
		addProps := sMap["additionalProperties"].(map[string]interface{})

		// Current implementation fails this assertion
		assert.Equal(t, "object", addProps["type"])
	})

	t.Run("OneOf Sanitization", func(t *testing.T) {
		input := map[string]interface{}{
			"oneOf": []interface{}{
				map[string]interface{}{
					// Missing type: object
					"properties": map[string]interface{}{
						"foo": map[string]interface{}{"type": "string"},
					},
				},
			},
		}

		sanitized, err := SanitizeJSONSchema(input)
		require.NoError(t, err)

		sMap := sanitized.AsMap()
		oneOf := sMap["oneOf"].([]interface{})
		item0 := oneOf[0].(map[string]interface{})

		// Current implementation fails this assertion
		assert.Equal(t, "object", item0["type"])
	})

	t.Run("Definitions Sanitization", func(t *testing.T) {
		input := map[string]interface{}{
			"$defs": map[string]interface{}{
				"MyType": map[string]interface{}{
					// Missing type: object
					"properties": map[string]interface{}{
						"foo": map[string]interface{}{"type": "string"},
					},
				},
			},
			"definitions": map[string]interface{}{
				"OldType": map[string]interface{}{
					// Missing type: object
					"properties": map[string]interface{}{
						"bar": map[string]interface{}{"type": "string"},
					},
				},
			},
		}

		sanitized, err := SanitizeJSONSchema(input)
		require.NoError(t, err)

		sMap := sanitized.AsMap()

		if defs, ok := sMap["$defs"].(map[string]interface{}); ok {
			myType := defs["MyType"].(map[string]interface{})
			assert.Equal(t, "object", myType["type"])
		} else {
			assert.Fail(t, "Expected $defs to be present and a map")
		}

		if definitions, ok := sMap["definitions"].(map[string]interface{}); ok {
			oldType := definitions["OldType"].(map[string]interface{})
			assert.Equal(t, "object", oldType["type"])
		} else {
			assert.Fail(t, "Expected definitions to be present and a map")
		}
	})

	t.Run("Invalid Types Graceful Handling", func(t *testing.T) {
		// Properties should be a map, but if it's not, we shouldn't crash
		input := map[string]interface{}{
			"type": "object",
			"properties": "invalid",
		}

		sanitized, err := SanitizeJSONSchema(input)
		require.NoError(t, err)

		sMap := sanitized.AsMap()
		assert.Equal(t, "invalid", sMap["properties"])
	})
}
