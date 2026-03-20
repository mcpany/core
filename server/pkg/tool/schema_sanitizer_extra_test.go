package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeJSONSchema_Nil(t *testing.T) {
	sanitized, err := SanitizeJSONSchema(nil)
	assert.NoError(t, err)
	assert.Nil(t, sanitized)
}

func TestSanitizeJSONSchema_MaxRecursionDepth(t *testing.T) {
	// Create a deeply nested schema that exceeds maxRecursionDepth (100)
	var root map[string]interface{}
	current := make(map[string]interface{})
	root = current

	for i := 0; i <= maxRecursionDepth+1; i++ {
		next := make(map[string]interface{})
		current["nested"] = next
		current = next
	}

	sanitized, err := SanitizeJSONSchema(root)
	assert.Error(t, err)
	assert.Nil(t, sanitized)
	assert.Contains(t, err.Error(), "exceeds maximum recursion depth")
}

func TestSanitizeJSONSchema_NotAMap(t *testing.T) {
	// If it's not a map, it should just convert it as is
	sanitized, err := SanitizeJSONSchema([]interface{}{"string"})
	// structpb.NewStruct expects a map, so it will fail when it tries to convert
	assert.Error(t, err)
	assert.Nil(t, sanitized)
}

func TestSanitizeJSONSchema_AllCombinators(t *testing.T) {
	rawSchema := map[string]interface{}{
		"items": map[string]interface{}{
			"properties": map[string]interface{}{
				"a": map[string]interface{}{"type": "string"},
			},
		},
		"additionalProperties": map[string]interface{}{
			"properties": map[string]interface{}{
				"b": map[string]interface{}{"type": "string"},
			},
		},
		"oneOf": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"c": map[string]interface{}{"type": "string"},
				},
			},
		},
		"anyOf": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"d": map[string]interface{}{"type": "string"},
				},
			},
		},
		"allOf": []interface{}{
			map[string]interface{}{
				"properties": map[string]interface{}{
					"e": map[string]interface{}{"type": "string"},
				},
			},
		},
		"definitions": map[string]interface{}{
			"def1": map[string]interface{}{
				"properties": map[string]interface{}{
					"f": map[string]interface{}{"type": "string"},
				},
			},
		},
		"$defs": map[string]interface{}{
			"def2": map[string]interface{}{
				"properties": map[string]interface{}{
					"g": map[string]interface{}{"type": "string"},
				},
			},
		},
	}

	sanitized, err := SanitizeJSONSchema(rawSchema)
	assert.NoError(t, err)
	assert.NotNil(t, sanitized)

	m := sanitized.AsMap()

	items := m["items"].(map[string]interface{})
	assert.Equal(t, "object", items["type"])

	addProps := m["additionalProperties"].(map[string]interface{})
	assert.Equal(t, "object", addProps["type"])

	oneOf := m["oneOf"].([]interface{})
	assert.Equal(t, "object", oneOf[0].(map[string]interface{})["type"])

	anyOf := m["anyOf"].([]interface{})
	assert.Equal(t, "object", anyOf[0].(map[string]interface{})["type"])

	allOf := m["allOf"].([]interface{})
	assert.Equal(t, "object", allOf[0].(map[string]interface{})["type"])

	definitions := m["definitions"].(map[string]interface{})
	assert.Equal(t, "object", definitions["def1"].(map[string]interface{})["type"])

	defs := m["$defs"].(map[string]interface{})
	assert.Equal(t, "object", defs["def2"].(map[string]interface{})["type"])
}
