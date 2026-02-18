// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"reflect"

	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/types/known/structpb"
)

// SanitizeJSONSchema attempts to fix common schema issues that cause strict MCP clients to fail.
// It takes a raw map[string]interface{} (or compatible) and returns a *structpb.Struct.
// This function does NOT modify the input schema.
func SanitizeJSONSchema(schema any) (*structpb.Struct, error) {
	if schema == nil {
		return nil, nil
	}

	// Deep copy schema to avoid mutation of the input, with cycle detection
	schemaCopy := deepCopyJSON(schema)

	// Sanitize in place, with cycle detection for the recursive calls
	return sanitizeJSONSchemaInPlace(schemaCopy, make(map[uintptr]bool))
}

func sanitizeJSONSchemaInPlace(schema any, visited map[uintptr]bool) (*structpb.Struct, error) {
	if schema == nil {
		return nil, nil
	}

	// Check for cycles using reflection
	val := reflect.ValueOf(schema)
	var ptr uintptr
	if val.Kind() == reflect.Map || val.Kind() == reflect.Slice || val.Kind() == reflect.Ptr {
		if !val.IsNil() {
			ptr = val.Pointer()
			if visited[ptr] {
				// Cycle detected, break recursion
				return nil, nil
			}
			visited[ptr] = true
			defer delete(visited, ptr)
		}
	}

	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		// If it's not a map, we can't really sanitize it easily, but it might be valid if it's not an object type.
		// However, for MCP InputSchema, it should be an object.
		// Let's try to convert it as is.
		return convertJSONSchemaToStruct(schema)
	}

	// 1. Fix missing "type": "object" if "properties" is present
	if _, hasProperties := schemaMap["properties"]; hasProperties {
		if _, hasType := schemaMap["type"]; !hasType {
			logging.GetLogger().Warn("Sanitizing schema: adding missing 'type': 'object' because 'properties' is present")
			schemaMap["type"] = "object"
		}
	}

	// Helper to sanitize a map of schemas (e.g. properties, definitions)
	sanitizeMapOfSchemas := func(fieldName string) {
		if props, ok := schemaMap[fieldName].(map[string]interface{}); ok {
			for k, v := range props {
				if vMap, ok := v.(map[string]interface{}); ok {
					// Recursively sanitize
					sanitizedV, err := sanitizeJSONSchemaInPlace(vMap, visited)
					if err == nil && sanitizedV != nil {
						props[k] = sanitizedV.AsMap()
					}
				}
			}
		}
	}

	// Helper to sanitize a field that can be a schema OR an array of schemas (e.g. items, allOf)
	sanitizeSchemaOrArrayOfSchemas := func(fieldName string) {
		if val, ok := schemaMap[fieldName]; ok {
			// Check for Array (e.g. items: [schema1, schema2] or allOf: [...])
			if schemaList, ok := val.([]interface{}); ok {
				newList := make([]interface{}, len(schemaList))
				for i, item := range schemaList {
					if itemMap, ok := item.(map[string]interface{}); ok {
						sanitized, err := sanitizeJSONSchemaInPlace(itemMap, visited)
						if err == nil && sanitized != nil {
							newList[i] = sanitized.AsMap()
						} else {
							newList[i] = item
						}
					} else {
						newList[i] = item
					}
				}
				schemaMap[fieldName] = newList
			} else if valMap, ok := val.(map[string]interface{}); ok {
				// Single schema object
				sanitized, err := sanitizeJSONSchemaInPlace(valMap, visited)
				if err == nil && sanitized != nil {
					schemaMap[fieldName] = sanitized.AsMap()
				}
			}
		}
	}

	// 2. Recursively sanitize properties
	sanitizeMapOfSchemas("properties")

	// 3. Recursively sanitize items (for arrays)
	sanitizeSchemaOrArrayOfSchemas("items")

	// 4. Handle additionalProperties (usually a schema)
	sanitizeSchemaOrArrayOfSchemas("additionalProperties")

	// 5. Handle allOf, anyOf, oneOf (arrays of schemas)
	sanitizeSchemaOrArrayOfSchemas("allOf")
	sanitizeSchemaOrArrayOfSchemas("anyOf")
	sanitizeSchemaOrArrayOfSchemas("oneOf")

	// 6. Handle definitions / $defs (map of schemas)
	sanitizeMapOfSchemas("definitions")
	sanitizeMapOfSchemas("$defs")

	// Let's convert back to structpb
	return structpb.NewStruct(schemaMap)
}

func deepCopyJSON(src any) any {
	return deepCopyJSONRecursive(src, make(map[uintptr]bool))
}

func deepCopyJSONRecursive(src any, visited map[uintptr]bool) any {
	if src == nil {
		return nil
	}

	// Check for cycles using reflection for Maps and Slices
	val := reflect.ValueOf(src)
	var ptr uintptr

	switch val.Kind() {
	case reflect.Map, reflect.Slice, reflect.Ptr:
		if !val.IsNil() {
			ptr = val.Pointer()
			if visited[ptr] {
				return nil // Cycle detected, return nil
			}
			visited[ptr] = true
			defer delete(visited, ptr)
		}
	}

	switch v := src.(type) {
	case map[string]interface{}:
		dst := make(map[string]interface{}, len(v))
		for k, val := range v {
			dst[k] = deepCopyJSONRecursive(val, visited)
		}
		return dst
	case []interface{}:
		dst := make([]interface{}, len(v))
		for i, val := range v {
			dst[i] = deepCopyJSONRecursive(val, visited)
		}
		return dst
	default:
		return v
	}
}
