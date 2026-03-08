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

func TestSanitizeSchema_NonMapTypes(t *testing.T) {
	// A non-map schema object that is handled by convertJSONSchemaToStruct
	rawSchema := "string_schema"

	// convertJSONSchemaToStruct returns an error for scalar string, "schema is not a valid JSON object"
	_, err := SanitizeJSONSchema(rawSchema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema is not a valid JSON object")
}

func TestSanitizeSchema_DeepCopyJSON(t *testing.T) {
	// Test depth limit
	rawSchema := map[string]interface{}{
		"type": "object",
	}

	// Create a deeply nested structure to hit maxRecursionDepth (100)
	current := rawSchema
	for i := 0; i < 105; i++ {
		next := map[string]interface{}{
			"type": "object",
		}
		current["properties"] = map[string]interface{}{
			"nested": next,
		}
		current = next
	}

	_, err := SanitizeJSONSchema(rawSchema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum recursion depth")
}

func TestSanitizeSchema_ItemsArrayAndObject(t *testing.T) {
	rawSchema := map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"properties": map[string]interface{}{
				"foo": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	items := sanitizedMap["items"].(map[string]interface{})
	assert.Equal(t, "object", items["type"])

	rawSchemaArray := map[string]interface{}{
		"type": "array",
		"items": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"foo": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}

	sanitizedArray, err := SanitizeJSONSchema(rawSchemaArray)
	assert.NoError(t, err)

	sanitizedMapArray := sanitizedArray.AsMap()
	itemsArray := sanitizedMapArray["items"].([]interface{})
	item0 := itemsArray[0].(map[string]interface{})
	assert.Equal(t, "object", item0["type"])
}

func TestSanitizeSchema_AdditionalProperties(t *testing.T) {
	rawSchema := map[string]interface{}{
		"type": "object",
		"additionalProperties": map[string]interface{}{
			"properties": map[string]interface{}{
				"foo": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	sanitizedMap := sanitized.AsMap()
	addProps := sanitizedMap["additionalProperties"].(map[string]interface{})
	assert.Equal(t, "object", addProps["type"])

	// Test boolean
	rawSchemaBool := map[string]interface{}{
		"type": "object",
		"additionalProperties": false,
	}

	sanitizedBool, err := SanitizeJSONSchema(rawSchemaBool)
	assert.NoError(t, err)

	sanitizedMapBool := sanitizedBool.AsMap()
	assert.Equal(t, false, sanitizedMapBool["additionalProperties"])
}

func TestSanitizeSchema_Combinators(t *testing.T) {
	rawSchema := map[string]interface{}{
		"oneOf": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"foo": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
		"anyOf": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"bar": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
		"allOf": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"baz": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)

	sanitizedMap := sanitized.AsMap()

	oneOf := sanitizedMap["oneOf"].([]interface{})
	assert.Equal(t, "object", oneOf[0].(map[string]interface{})["type"])

	anyOf := sanitizedMap["anyOf"].([]interface{})
	assert.Equal(t, "object", anyOf[0].(map[string]interface{})["type"])

	allOf := sanitizedMap["allOf"].([]interface{})
	assert.Equal(t, "object", allOf[0].(map[string]interface{})["type"])
}

func TestSanitizeSchema_Definitions(t *testing.T) {
	rawSchema := map[string]interface{}{
		"definitions": map[string]interface{}{
			"def1": map[string]interface{}{
				"properties": map[string]interface{}{
					"foo": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
		"$defs": map[string]interface{}{
			"def2": map[string]interface{}{
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

	defs1 := sanitizedMap["definitions"].(map[string]interface{})
	assert.Equal(t, "object", defs1["def1"].(map[string]interface{})["type"])

	defs2 := sanitizedMap["$defs"].(map[string]interface{})
	assert.Equal(t, "object", defs2["def2"].(map[string]interface{})["type"])
}

func TestSanitizeSchema_DeepCopyJSON_ArrayDeep(t *testing.T) {
	// Test depth limit for deepCopyJSON array
	rawSchema := []interface{}{
		"item",
	}

	// Create a deeply nested array structure to hit maxRecursionDepth (100)
	current := rawSchema
	for i := 0; i < 105; i++ {
		next := []interface{}{"item"}
		current[0] = next
		current = next
	}

	// This should fail early in deepCopyJSON
	_, err := SanitizeJSONSchema(rawSchema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum recursion depth")
}

func TestSanitizeSchema_DeepCopyJSON_MapArrayMixed(t *testing.T) {
	rawSchema := map[string]interface{}{
		"list": []interface{}{
			map[string]interface{}{
				"key": "value",
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)
	assert.NotNil(t, sanitized)
}

func TestSanitizeSchema_Nil(t *testing.T) {
	sanitized, err := SanitizeJSONSchema(nil)
	assert.NoError(t, err)
	assert.Nil(t, sanitized)
}

func TestSanitizeSchema_Recurse_ErrorPropagation(t *testing.T) {
    // We cannot easily test the recursion limit in sanitizeJSONSchemaInPlace from the public SanitizeJSONSchema API
	// because deepCopyJSON uses the SAME maxRecursionDepth limit and will hit it first.
	// But we can test it by manually creating an alias to the private function via reflection or just knowing we covered deepCopyJSON.
	// Let's call the unexported sanitizeJSONSchemaInPlace directly from the test since they are in the same package `tool`.

    rawSchema := map[string]interface{}{
        "properties": map[string]interface{}{
            "level1": map[string]interface{}{
                "properties": map[string]interface{}{
                    "level2": map[string]interface{}{
                        "type": "string",
                    },
                },
            },
        },
    }

    // Depth 99 -> properties gets depth 100 -> level1 properties gets 101 (error)
    _, err := sanitizeJSONSchemaInPlace(rawSchema, 99)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "exceeds maximum recursion depth")

    // Test items error
    rawSchemaItems := map[string]interface{}{
        "items": map[string]interface{}{
            "items": map[string]interface{}{
                "type": "string",
            },
        },
    }
    _, errItems := sanitizeJSONSchemaInPlace(rawSchemaItems, 99)
    assert.Error(t, errItems)

    // Test list items error
    rawSchemaListItems := map[string]interface{}{
        "items": []interface{}{
            map[string]interface{}{
                "type": "string",
            },
        },
    }
    _, errListItems := sanitizeJSONSchemaInPlace(rawSchemaListItems, 100)
    assert.Error(t, errListItems)

    // Test additionalProperties error
    rawSchemaAdd := map[string]interface{}{
        "additionalProperties": map[string]interface{}{
            "type": "string",
        },
    }
    _, errAdd := sanitizeJSONSchemaInPlace(rawSchemaAdd, 100)
    assert.Error(t, errAdd)

    // Test combinators error
    rawSchemaComb := map[string]interface{}{
        "oneOf": []interface{}{
            map[string]interface{}{
                "type": "string",
            },
        },
    }
    _, errComb := sanitizeJSONSchemaInPlace(rawSchemaComb, 100)
    assert.Error(t, errComb)

    // Test definitions error
    rawSchemaDef := map[string]interface{}{
        "definitions": map[string]interface{}{
            "def1": map[string]interface{}{
                "type": "string",
            },
        },
    }
    _, errDef := sanitizeJSONSchemaInPlace(rawSchemaDef, 100)
    assert.Error(t, errDef)
}

func TestSanitizeSchema_InvalidMapKey(t *testing.T) {
	// A map containing an invalid key/value that cannot be converted to structpb
	rawSchema := map[string]interface{}{
		"properties": map[string]interface{}{
			"foo": make(chan int),
		},
	}

	// This will cause an error inside sanitizeJSONSchemaInPlace when calling structpb.NewStruct
	_, err := SanitizeJSONSchema(rawSchema)
	assert.Error(t, err)
}
