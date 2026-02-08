// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"

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

	// Deep copy schema to avoid mutation of the input, with depth limit to prevent stack overflow
	schemaCopy, err := deepCopyJSON(schema, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to deep copy schema: %w", err)
	}

	return sanitizeJSONSchemaInPlace(schemaCopy, 0)
}

const maxSchemaDepth = 100

func sanitizeJSONSchemaInPlace(schema any, depth int) (*structpb.Struct, error) {
	if depth > maxSchemaDepth {
		return nil, fmt.Errorf("schema too deep")
	}

	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		// If it's not a map, we can't really sanitize it easily.
		// It might be valid if it's not an object type (e.g. boolean schema in newer drafts).
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
				sanitizedV, err := sanitizeJSONSchemaInPlace(vMap, depth+1)
				if err == nil && sanitizedV != nil {
					props[k] = sanitizedV.AsMap()
				}
			}
		}
	}

	// 3. Recursively sanitize items (for arrays)
	// Handle items as object (single schema)
	if itemsMap, ok := schemaMap["items"].(map[string]interface{}); ok {
		sanitizedItems, err := sanitizeJSONSchemaInPlace(itemsMap, depth+1)
		if err == nil && sanitizedItems != nil {
			schemaMap["items"] = sanitizedItems.AsMap()
		}
	} else if itemsSlice, ok := schemaMap["items"].([]interface{}); ok {
		// Handle items as array (tuple validation)
		sanitizedSlice := make([]interface{}, len(itemsSlice))
		for i, item := range itemsSlice {
			if itemMap, ok := item.(map[string]interface{}); ok {
				sanitizedItem, err := sanitizeJSONSchemaInPlace(itemMap, depth+1)
				if err == nil && sanitizedItem != nil {
					sanitizedSlice[i] = sanitizedItem.AsMap()
				} else {
					sanitizedSlice[i] = item
				}
			} else {
				sanitizedSlice[i] = item
			}
		}
		schemaMap["items"] = sanitizedSlice
	}

	// 4. Handle additionalProperties (schema or boolean)
	if additionalProps, ok := schemaMap["additionalProperties"].(map[string]interface{}); ok {
		sanitizedAddProps, err := sanitizeJSONSchemaInPlace(additionalProps, depth+1)
		if err == nil && sanitizedAddProps != nil {
			schemaMap["additionalProperties"] = sanitizedAddProps.AsMap()
		}
	}

	// Let's convert back to structpb
	return structpb.NewStruct(schemaMap)
}

func deepCopyJSON(src any, depth int) (any, error) {
	if depth > maxSchemaDepth {
		return nil, fmt.Errorf("recursion limit reached")
	}

	switch v := src.(type) {
	case map[string]interface{}:
		dst := make(map[string]interface{}, len(v))
		for k, val := range v {
			copyVal, err := deepCopyJSON(val, depth+1)
			if err != nil {
				return nil, err
			}
			dst[k] = copyVal
		}
		return dst, nil
	case []interface{}:
		dst := make([]interface{}, len(v))
		for i, val := range v {
			copyVal, err := deepCopyJSON(val, depth+1)
			if err != nil {
				return nil, err
			}
			dst[i] = copyVal
		}
		return dst, nil
	default:
		return v, nil
	}
}
