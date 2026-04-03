// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"fmt"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

var (
	schemaOnce     sync.Once
	compiledSchema *jsonschema.Schema
	schemaGenErr   error
)

// ensureSchema generates and compiles the JSON schema for McpAnyServerConfig.
// It does this only once.
func ensureSchema() (*jsonschema.Schema, error) {
	schemaOnce.Do(func() {
		// 1. Generate JSON Schema from the Proto definition
		cfg := configv1.McpAnyServerConfig_builder{}.Build()
		var err error
		compiledSchema, err = GenerateSchemaFromProto(cfg.ProtoReflect())
		if err != nil {
			schemaGenErr = fmt.Errorf("failed to generate schema from proto: %w", err)
			return
		}
	})
	return compiledSchema, schemaGenErr
}

// ValidateConfigAgainstSchema serves as a public interface for interacting with ValidateConfigAgainstSchema.
//
// Summary: Validate the config against schema appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func ValidateConfigAgainstSchema(rawConfig map[string]interface{}) error {
	schema, err := ensureSchema()
	if err != nil {
		return fmt.Errorf("schema generation failed: %w", err)
	}

	if err := schema.Validate(rawConfig); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}
	return nil
}

// GenerateJSONSchemaBytes serves as a public interface for interacting with GenerateJSONSchemaBytes.
//
// Summary: Generate the json schema bytes appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func GenerateJSONSchemaBytes() ([]byte, error) {
	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	schemaMap := GenerateSchemaMapFromProto(cfg.ProtoReflect())
	return json.MarshalIndent(schemaMap, "", "  ")
}
