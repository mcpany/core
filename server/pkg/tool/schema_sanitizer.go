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
// ⚡ BOLT: Optimized to single-pass traversal for deep copy, sanitization, and structpb conversion.
// Randomized Selection from Top 5 High-Impact Targets.
func SanitizeJSONSchema(schema any) (*structpb.Struct, error) {
	if schema == nil {
		return nil, nil
	}

	val, err := convertAndSanitize(schema)
	if err != nil {
		return nil, err
	}

	// Ensure it is a Struct at the top level
	if s, ok := val.Kind.(*structpb.Value_StructValue); ok {
		return s.StructValue, nil
	}

	// If the top level is not a struct, try to convert via structpb directly to see if it handles it (e.g. error msg)
	// or return a specific error.
	return nil, fmt.Errorf("schema must be a JSON object, got %T", schema)
}

func convertAndSanitize(v any) (*structpb.Value, error) {
	switch val := v.(type) {
	case map[string]interface{}:
		fields := make(map[string]*structpb.Value, len(val))
		for k, v := range val {
			sv, err := convertAndSanitize(v)
			if err != nil {
				return nil, err
			}
			fields[k] = sv
		}

		// Sanitize: Check for properties without type: object
		// This applies to the root schema, nested properties, items, definitions, etc.
		if _, hasProperties := fields["properties"]; hasProperties {
			if _, hasType := fields["type"]; !hasType {
				logging.GetLogger().Warn("Sanitizing schema: adding missing 'type': 'object' because 'properties' is present")
				fields["type"] = structpb.NewStringValue("object")
			}
		}

		return &structpb.Value{Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: fields}}}, nil

	case []interface{}:
		list := make([]*structpb.Value, len(val))
		for i, v := range val {
			sv, err := convertAndSanitize(v)
			if err != nil {
				return nil, err
			}
			list[i] = sv
		}
		return &structpb.Value{Kind: &structpb.Value_ListValue{ListValue: &structpb.ListValue{Values: list}}}, nil

	default:
		// Use structpb.NewValue for primitives (it handles bool, string, number, nil)
		return structpb.NewValue(val)
	}
}
