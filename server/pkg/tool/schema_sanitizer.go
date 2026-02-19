// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
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

	// Deep copy schema to avoid mutation of the input
	schemaCopy := deepCopyJSON(schema)

	// Since we prune cycles in deepCopyJSON, sanitizeJSONSchemaInPlace is safe
	return sanitizeJSONSchemaInPlace(schemaCopy)
}

const maxSchemaDepth = 500

func sanitizeJSONSchemaInPlace(schema any) (*structpb.Struct, error) {
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

	// 2. Recursively sanitize properties
	if props, ok := schemaMap["properties"].(map[string]interface{}); ok {
		for k, v := range props {
			if vMap, ok := v.(map[string]interface{}); ok {
				// Recursively sanitize
				sanitizedV, err := sanitizeJSONSchemaInPlace(vMap)
				if err == nil && sanitizedV != nil {
					props[k] = sanitizedV.AsMap()
				}
			}
		}
	}

	// 3. Recursively sanitize items (for arrays)
	if items, ok := schemaMap["items"].(map[string]interface{}); ok {
		sanitizedItems, err := sanitizeJSONSchemaInPlace(items)
		if err == nil && sanitizedItems != nil {
			schemaMap["items"] = sanitizedItems.AsMap()
		}
	} else if itemsArr, ok := schemaMap["items"].([]interface{}); ok {
		// Handle tuple validation (items: [schema1, schema2])
		for i, item := range itemsArr {
			if itemMap, ok := item.(map[string]interface{}); ok {
				sanitizedItem, err := sanitizeJSONSchemaInPlace(itemMap)
				if err == nil && sanitizedItem != nil {
					itemsArr[i] = sanitizedItem.AsMap()
				}
			}
		}
	}

	// 4. Sanitize additionalProperties
	if addProps, ok := schemaMap["additionalProperties"].(map[string]interface{}); ok {
		sanitized, err := sanitizeJSONSchemaInPlace(addProps)
		if err == nil && sanitized != nil {
			schemaMap["additionalProperties"] = sanitized.AsMap()
		}
	}

	// 5. Sanitize oneOf, anyOf, allOf
	for _, key := range []string{"oneOf", "anyOf", "allOf"} {
		if arr, ok := schemaMap[key].([]interface{}); ok {
			for i, item := range arr {
				if itemMap, ok := item.(map[string]interface{}); ok {
					sanitized, err := sanitizeJSONSchemaInPlace(itemMap)
					if err == nil && sanitized != nil {
						arr[i] = sanitized.AsMap()
					}
				}
			}
		}
	}

	// 6. Sanitize $defs and definitions
	for _, key := range []string{"$defs", "definitions"} {
		if defs, ok := schemaMap[key].(map[string]interface{}); ok {
			for k, v := range defs {
				if vMap, ok := v.(map[string]interface{}); ok {
					sanitized, err := sanitizeJSONSchemaInPlace(vMap)
					if err == nil && sanitized != nil {
						defs[k] = sanitized.AsMap()
					}
				}
			}
		}
	}

	// Let's convert back to structpb
	return structpb.NewStruct(schemaMap)
}

func deepCopyJSON(src any) any {
	return deepCopyJSONWithLimit(src, 0)
}

func deepCopyJSONWithLimit(src any, depth int) any {
	if depth > maxSchemaDepth {
		// Prune deep branches/cycles to avoid stack overflow
		return nil
	}

	switch v := src.(type) {
	case map[string]interface{}:
		dst := make(map[string]interface{}, len(v))
		for k, val := range v {
			dst[k] = deepCopyJSONWithLimit(val, depth+1)
		}
		return dst
	case []interface{}:
		dst := make([]interface{}, len(v))
		for i, val := range v {
			dst[i] = deepCopyJSONWithLimit(val, depth+1)
		}
		return dst
	default:
		return v
	}
}
