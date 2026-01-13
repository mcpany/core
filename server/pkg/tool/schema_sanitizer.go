// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/types/known/structpb"
)

// SanitizeJSONSchema attempts to fix common schema issues that cause strict MCP clients to fail.
// It takes a raw map[string]interface{} (or compatible) and returns a *structpb.Struct.
func SanitizeJSONSchema(schema any) (*structpb.Struct, error) {
	if schema == nil {
		return nil, nil
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

	// 2. Recursively sanitize properties
	if props, ok := schemaMap["properties"].(map[string]interface{}); ok {
		for k, v := range props {
			if vMap, ok := v.(map[string]interface{}); ok {
				// Recursively sanitize
				sanitizedV, err := SanitizeJSONSchema(vMap)
				if err == nil && sanitizedV != nil {
					props[k] = sanitizedV.AsMap()
				}
			}
		}
	}

	// 3. Handle top-level oneOf/anyOf/allOf if they are not supported by some clients?
	// The issue #10606 says: "input_schema does not support oneOf, allOf, or anyOf at the top level"
	// If we detect this at the top level, what can we do?
	// We could try to merge them if possible, or just pick the first one?
	// Merging is hard. Picking the first one is lossy.
	// For now, let's just log a warning if we see it at the top level.
	// Note: We need to know if this is the "top level" of the Tool's InputSchema.
	// Since this function is recursive, we might need a context or just assume the caller handles the top-level check.
	// But `SanitizeJSONSchema` is likely called on the root schema.

	// Let's convert back to structpb
	return structpb.NewStruct(schemaMap)
}
