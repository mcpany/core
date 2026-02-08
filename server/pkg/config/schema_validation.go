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

// ValidateConfigAgainstSchema validates the raw configuration map against the generated JSON schema.
//
// Summary: validates the raw configuration map against the generated JSON schema.
//
// Parameters:
//   - rawConfig: map[string]interface{}. The rawConfig.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// GenerateJSONSchemaBytes returns the JSON schema for McpAnyServerConfig as a byte slice.
//
// Summary: returns the JSON schema for McpAnyServerConfig as a byte slice.
//
// Parameters:
//   None.
//
// Returns:
//   - []byte: The []byte.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func GenerateJSONSchemaBytes() ([]byte, error) {
	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	schemaMap := GenerateSchemaMapFromProto(cfg.ProtoReflect())
	return json.MarshalIndent(schemaMap, "", "  ")
}
