// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"

	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/types/known/structpb"
)

const maxRecursionDepth = 100

// SanitizeJSONSchema attempts to fix common schema issues that cause strict MCP clients to fail.
// It takes a raw map[string]interface{} (or compatible) and returns a *structpb.Struct.
// This function does NOT modify the input schema.
func SanitizeJSONSchema(schema any) (*structpb.Struct, error) {
	if schema == nil {
		return nil, nil
	}

	// Deep copy schema to avoid mutation of the input
	// We use a depth-limited deep copy now to prevent stack overflow on circular refs
	schemaCopy, err := deepCopyJSON(schema, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to copy schema: %w", err)
	}

	return sanitizeJSONSchemaInPlace(schemaCopy, 0)
}

func sanitizeJSONSchemaInPlace(schema any, depth int) (*structpb.Struct, error) {
	if depth > maxRecursionDepth {
		return nil, fmt.Errorf("schema exceeds maximum recursion depth of %d", maxRecursionDepth)
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

	// Helper to recurse
	var recurse func(v any) (any, error)
	recurse = func(v any) (any, error) {
		// If it's a schema object (map), sanitize it
		if m, ok := v.(map[string]interface{}); ok {
			res, err := sanitizeJSONSchemaInPlace(m, depth+1)
			if err != nil {
				return nil, err
			}
			return res.AsMap(), nil
		}
		// If it's a list of schemas (e.g. oneOf), sanitize each
		if list, ok := v.([]interface{}); ok {
			newList := make([]interface{}, len(list))
			for i, item := range list {
				res, err := recurse(item)
				if err != nil {
					return nil, err
				}
				newList[i] = res
			}
			return newList, nil
		}
		return v, nil
	}

	// 2. Recursively sanitize properties
	if props, ok := schemaMap["properties"].(map[string]interface{}); ok {
		for k, v := range props {
			res, err := recurse(v)
			if err != nil {
				return nil, err
			}
			props[k] = res
		}
	}

	// 3. Recursively sanitize items (for arrays)
	// Can be a single schema or an array of schemas
	if items, ok := schemaMap["items"]; ok {
		res, err := recurse(items)
		if err != nil {
			return nil, err
		}
		schemaMap["items"] = res
	}

	// 4. Handle additionalProperties (can be boolean or schema)
	if addProps, ok := schemaMap["additionalProperties"]; ok {
		// Only recurse if it's a map (schema), boolean is fine as is
		if _, isMap := addProps.(map[string]interface{}); isMap {
			res, err := recurse(addProps)
			if err != nil {
				return nil, err
			}
			schemaMap["additionalProperties"] = res
		}
	}

	// 5. Handle combinators (oneOf, anyOf, allOf)
	for _, key := range []string{"oneOf", "anyOf", "allOf"} {
		if val, ok := schemaMap[key]; ok {
			res, err := recurse(val)
			if err != nil {
				return nil, err
			}
			schemaMap[key] = res
		}
	}

	// 6. Handle definitions / $defs
	for _, key := range []string{"definitions", "$defs"} {
		if defs, ok := schemaMap[key].(map[string]interface{}); ok {
			for k, v := range defs {
				res, err := recurse(v)
				if err != nil {
					return nil, err
				}
				defs[k] = res
			}
		}
	}

	// Let's convert back to structpb
	return structpb.NewStruct(schemaMap)
}

func deepCopyJSON(src any, depth int) (any, error) {
	if depth > maxRecursionDepth {
		return nil, fmt.Errorf("deep copy exceeds maximum recursion depth of %d", maxRecursionDepth)
	}
	switch v := src.(type) {
	case map[string]interface{}:
		dst := make(map[string]interface{}, len(v))
		for k, val := range v {
			res, err := deepCopyJSON(val, depth+1)
			if err != nil {
				return nil, err
			}
			dst[k] = res
		}
		return dst, nil
	case []interface{}:
		dst := make([]interface{}, len(v))
		for i, val := range v {
			res, err := deepCopyJSON(val, depth+1)
			if err != nil {
				return nil, err
			}
			dst[i] = res
		}
		return dst, nil
	default:
		return v, nil
	}
}
