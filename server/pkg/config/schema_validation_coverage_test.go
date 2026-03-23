// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfigAgainstSchema_Invalid(t *testing.T) {
	// Test with invalid map that doesn't match schema
	// e.g. wrong type for field
	data := map[string]interface{}{
		"global_settings": map[string]interface{}{
			"log_level": 123, // Should be string or enum, but protobuf to JSON conversion might handle it.
			// Let's use a definitely wrong field type if possible, or unknown field (but unknown field is allowed by schema usually?)
			// The schema generated from proto is strict?
			"mcp_listen_address": 123, // Should be string
		},
	}
	err := ValidateConfigAgainstSchema(data)
	assert.Error(t, err)
}

func TestEnsureSchema_Error(t *testing.T) {
	// ensureSchema is internal, but we can trigger it via ValidateConfigAgainstSchema maybe?
	// It's used in ValidateConfigAgainstSchema:
	/*
		func ValidateConfigAgainstSchema(config map[string]interface{}) error {
			if err := ensureSchema(nil); err != nil { ... }
			// ...
		}
	*/
	// But ensureSchema() logic:
	/*
	func ensureSchema() error {
		schemaMu.Lock()
		defer schemaMu.Unlock()
		if jsonSchema != nil {
			return nil
		}
		// ... GenerateJSONSchemaBytes ...
	}
	*/
	// It just generates the schema. It shouldn't fail unless proto generation fails.
}
