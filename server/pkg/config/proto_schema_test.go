// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

// TestGenerateSchemaMapFromProto ...
// Summary: TestGenerateSchemaMapFromProto
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	schemaMap := GenerateSchemaMapFromProto(cfg.ProtoReflect())
	assert.NotNil(t, schemaMap)
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", schemaMap["$schema"])
	assert.NotNil(t, schemaMap["$defs"])
}

// TestCompileSchema_Error ...
// Summary: TestCompileSchema_Error
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Function values cannot be marshaled to JSON
	badMap := map[string]interface{}{
		"bad": func() {},
	}

	schema, err := CompileSchema(badMap)
	assert.Error(t, err)
	assert.Nil(t, schema)
	assert.Contains(t, err.Error(), "json: unsupported type")
}

// TestCompileSchema_InvalidSchema ...
// Summary: TestCompileSchema_InvalidSchema
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Valid JSON but invalid schema ($id must be a string)
	// This triggers AddResource parsing failure
	badSchema := map[string]interface{}{
		"$id": 123,
	}

	schema, err := CompileSchema(badSchema)
	assert.Error(t, err)
	assert.Nil(t, schema)
	// This usually returns "json: cannot unmarshal number into Go struct field Schema.id of type string" or similar
}
