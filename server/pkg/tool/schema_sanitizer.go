// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"

	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/types/known/structpb"
)

// maxSchemaDepth is the maximum depth for schema traversal to prevent stack overflows.
const maxSchemaDepth = 100

// SanitizeJSONSchema attempts to fix common schema issues that cause strict MCP clients to fail.
// It takes a raw map[string]interface{} (or compatible) and returns a *structpb.Struct.
// This function does NOT modify the input schema.
func SanitizeJSONSchema(schema any) (*structpb.Struct, error) {
	if schema == nil {
		return nil, nil
	}

	// Deep copy schema to avoid mutation of the input and break cycles
	schemaCopy := deepCopyJSON(schema, 0)
	if schemaCopy == nil {
		// If deep copy failed (e.g. only depth limit hit), we might return empty struct or error.
		// For now, let's assume it returned a truncated structure.
		// If the root was truncated, it's nil.
		return nil, fmt.Errorf("schema too deep or invalid")
	}

	return sanitizeJSONSchemaInPlace(schemaCopy, 0)
}

func sanitizeJSONSchemaInPlace(schema any, depth int) (*structpb.Struct, error) {
	if depth > maxSchemaDepth {
		return nil, fmt.Errorf("schema depth limit exceeded")
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

	// Helper to recurse safely
	recurse := func(v interface{}) (interface{}, error) {
		if vMap, ok := v.(map[string]interface{}); ok {
			sanitized, err := sanitizeJSONSchemaInPlace(vMap, depth+1)
			if err != nil {
				return nil, err
			}
			if sanitized != nil {
				return sanitized.AsMap(), nil
			}
		}
		return v, nil
	}

	// 2. Recursively sanitize properties
	if props, ok := schemaMap["properties"].(map[string]interface{}); ok {
		for k, v := range props {
			sanitizedV, err := recurse(v)
			if err == nil {
				props[k] = sanitizedV
			}
		}
	}

	// 3. Recursively sanitize items (for arrays)
	if items, ok := schemaMap["items"]; ok {
		if itemsMap, ok := items.(map[string]interface{}); ok {
			// Single schema
			sanitizedItems, err := sanitizeJSONSchemaInPlace(itemsMap, depth+1)
			if err == nil && sanitizedItems != nil {
				schemaMap["items"] = sanitizedItems.AsMap()
			}
		} else if itemsList, ok := items.([]interface{}); ok {
			// Array of schemas (tuple validation)
			newItems := make([]interface{}, len(itemsList))
			for i, item := range itemsList {
				if itemMap, ok := item.(map[string]interface{}); ok {
					sanitizedItem, err := sanitizeJSONSchemaInPlace(itemMap, depth+1)
					if err == nil && sanitizedItem != nil {
						newItems[i] = sanitizedItem.AsMap()
					} else {
						newItems[i] = item
					}
				} else {
					newItems[i] = item
				}
			}
			schemaMap["items"] = newItems
		}
	}

	// 4. Recursively sanitize additionalProperties
	if additionalProps, ok := schemaMap["additionalProperties"].(map[string]interface{}); ok {
		sanitizedAP, err := sanitizeJSONSchemaInPlace(additionalProps, depth+1)
		if err == nil && sanitizedAP != nil {
			schemaMap["additionalProperties"] = sanitizedAP.AsMap()
		}
	}

	// 5. Recursively sanitize $defs / definitions
	for _, key := range []string{"$defs", "definitions"} {
		if defs, ok := schemaMap[key].(map[string]interface{}); ok {
			for k, v := range defs {
				sanitizedV, err := recurse(v)
				if err == nil {
					defs[k] = sanitizedV
				}
			}
		}
	}

	// 6. Recursively sanitize oneOf, anyOf, allOf
	for _, key := range []string{"oneOf", "anyOf", "allOf"} {
		if list, ok := schemaMap[key].([]interface{}); ok {
			newList := make([]interface{}, len(list))
			for i, item := range list {
				if itemMap, ok := item.(map[string]interface{}); ok {
					sanitizedItem, err := sanitizeJSONSchemaInPlace(itemMap, depth+1)
					if err == nil && sanitizedItem != nil {
						newList[i] = sanitizedItem.AsMap()
					} else {
						newList[i] = item
					}
				} else {
					newList[i] = item
				}
			}
			schemaMap[key] = newList
		}
	}

	// Let's convert back to structpb
	return structpb.NewStruct(schemaMap)
}

func deepCopyJSON(src any, depth int) any {
	if depth > maxSchemaDepth {
		return nil
	}
	switch v := src.(type) {
	case map[string]interface{}:
		dst := make(map[string]interface{}, len(v))
		for k, val := range v {
			dst[k] = deepCopyJSON(val, depth+1)
		}
		return dst
	case []interface{}:
		dst := make([]interface{}, len(v))
		for i, val := range v {
			dst[i] = deepCopyJSON(val, depth+1)
		}
		return dst
	default:
		return v
	}
}
