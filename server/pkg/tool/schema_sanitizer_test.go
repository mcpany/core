// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeSchema_MissingTypeObject(t *testing.T) {
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

func TestSanitizeJSONSchema_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		schema      any
		wantSchema  map[string]interface{}
		wantErr     string
	}{
		{
			name:   "nil schema",
			schema: nil,
			wantSchema: nil,
		},
		{
			name:   "non-map scalar",
			schema: "this is a string",
			wantSchema: nil, // Note: convertJSONSchemaToStruct might not handle scalar string directly or will return an error, but let's see. Wait, convertJSONSchemaToStruct will wrap it or error. Wait, convertJSONSchemaToStruct returns structpb.NewStruct, which expects map.
		},
		{
			name: "items array",
			schema: map[string]interface{}{
				"type": "array",
				"items": []interface{}{
					map[string]interface{}{
						"properties": map[string]interface{}{},
					},
				},
			},
			wantSchema: map[string]interface{}{
				"type": "array",
				"items": []interface{}{
					map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{},
					},
				},
			},
		},
		{
			name: "combinators",
			schema: map[string]interface{}{
				"oneOf": []interface{}{
					map[string]interface{}{"properties": map[string]interface{}{}},
				},
				"anyOf": []interface{}{
					map[string]interface{}{"properties": map[string]interface{}{}},
				},
				"allOf": []interface{}{
					map[string]interface{}{"properties": map[string]interface{}{}},
				},
			},
			wantSchema: map[string]interface{}{
				"oneOf": []interface{}{
					map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
				},
				"anyOf": []interface{}{
					map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
				},
				"allOf": []interface{}{
					map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
				},
			},
		},
		{
			name: "definitions and defs",
			schema: map[string]interface{}{
				"definitions": map[string]interface{}{
					"a": map[string]interface{}{"properties": map[string]interface{}{}},
				},
				"$defs": map[string]interface{}{
					"b": map[string]interface{}{"properties": map[string]interface{}{}},
				},
			},
			wantSchema: map[string]interface{}{
				"definitions": map[string]interface{}{
					"a": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
				},
				"$defs": map[string]interface{}{
					"b": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
				},
			},
		},
		{
			name: "additionalProperties object",
			schema: map[string]interface{}{
				"additionalProperties": map[string]interface{}{
					"properties": map[string]interface{}{},
				},
			},
			wantSchema: map[string]interface{}{
				"additionalProperties": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			name: "additionalProperties boolean",
			schema: map[string]interface{}{
				"additionalProperties": true,
			},
			wantSchema: map[string]interface{}{
				"additionalProperties": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeJSONSchema(tt.schema)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			if tt.name == "non-map scalar" {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.wantSchema == nil {
				assert.Nil(t, got)
				return
			}

			assert.Equal(t, tt.wantSchema, got.AsMap())
		})
	}
}

func TestSanitizeJSONSchema_RecursionLimit(t *testing.T) {
	// Let's create a very deep map
	deepMap := map[string]interface{}{}
	current := deepMap
	for i := 0; i < 110; i++ {
		next := map[string]interface{}{}
		current["prop"] = next
		current = next
	}

	_, err := SanitizeJSONSchema(deepMap)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum recursion depth")
}

func TestSanitizeJSONSchema_DeepCopyRecursionLimit(t *testing.T) {
	// Actually deepCopyJSON fails first when we pass it to SanitizeJSONSchema
	// Let's test the inner deepCopy directly or just through SanitizeJSONSchema
	deepArray := []interface{}{}
	current := deepArray
	for i := 0; i < 110; i++ {
		next := []interface{}{}
		if i == 0 {
			deepArray = append(deepArray, next)
			current = next
		} else {
			current = append(current, next)
			current = next
		}
	}
	// Note: It's hard to make a deep array structure like this mutable easily without pointer loops.
	// We'll just test a recursive struct loop to trigger the depth.

	cyclicMap := map[string]interface{}{}
	cyclicMap["self"] = cyclicMap

	_, err := SanitizeJSONSchema(cyclicMap)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum recursion depth")
}
