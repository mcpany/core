// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGenerateSchemaMapFromProto(t *testing.T) {
	cfg := configv1.McpAnyServerConfig_builder{}.Build()
	schemaMap := GenerateSchemaMapFromProto(cfg.ProtoReflect())
	assert.NotNil(t, schemaMap)
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", schemaMap["$schema"])
	assert.NotNil(t, schemaMap["$defs"])
}

func TestGenerateSchemaFromProto_Complex(t *testing.T) {
	cfg := &configv1.McpAnyServerConfig{}
	schemaMap := GenerateSchemaMapFromProto(cfg.ProtoReflect())

	assert.NotNil(t, schemaMap)
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", schemaMap["$schema"])

	defs, ok := schemaMap["$defs"].(map[string]interface{})
	assert.True(t, ok, "$defs should be a map")

	// Check if key definitions exist
	assert.Contains(t, defs, "mcpany.config.v1.McpAnyServerConfig")
	assert.Contains(t, defs, "mcpany.config.v1.GlobalSettings")
	assert.Contains(t, defs, "mcpany.config.v1.UpstreamServiceConfig")

	// Check root reference
	rootRef, ok := schemaMap["$ref"].(string)
	assert.True(t, ok, "$ref should be a string")
	assert.Equal(t, "#/$defs/mcpany.config.v1.McpAnyServerConfig", rootRef)

	// Inspect McpAnyServerConfig definition
	rootDef := defs["mcpany.config.v1.McpAnyServerConfig"].(map[string]interface{})
	assert.Equal(t, "object", rootDef["type"])
	props := rootDef["properties"].(map[string]interface{})

	// Check fields
	assert.Contains(t, props, "global_settings")
	assert.Contains(t, props, "upstream_services")
	assert.Contains(t, props, "collections")

	// Check global_settings ref
	gs := props["global_settings"].(map[string]interface{})
	assert.Equal(t, "#/$defs/mcpany.config.v1.GlobalSettings", gs["$ref"])

	// Check upstream_services array
	us := props["upstream_services"].(map[string]interface{})
	assert.Equal(t, "array", us["type"])
	items := us["items"].(map[string]interface{})
	assert.Equal(t, "#/$defs/mcpany.config.v1.UpstreamServiceConfig", items["$ref"])
}

func TestCompileSchema_Error(t *testing.T) {
	// Function values cannot be marshaled to JSON
	badMap := map[string]interface{}{
		"bad": func() {},
	}

	schema, err := CompileSchema(badMap)
	assert.Error(t, err)
	assert.Nil(t, schema)
	assert.Contains(t, err.Error(), "json: unsupported type")
}

func TestCompileSchema_InvalidSchema(t *testing.T) {
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

func TestGenerateSchemaFromProto_Types(t *testing.T) {
	// Test Duration
	dur := durationpb.New(0)
	schemaMap := GenerateSchemaMapFromProto(dur.ProtoReflect())
	defs := schemaMap["$defs"].(map[string]interface{})
	durDef := defs["google.protobuf.Duration"].(map[string]interface{})

	assert.Equal(t, "string", durDef["type"])
	assert.Contains(t, durDef, "pattern")

	// Test Struct
	st := &structpb.Struct{}
	schemaMap = GenerateSchemaMapFromProto(st.ProtoReflect())
	defs = schemaMap["$defs"].(map[string]interface{})
	stDef := defs["google.protobuf.Struct"].(map[string]interface{})

	assert.Equal(t, "object", stDef["type"])
	assert.Equal(t, true, stDef["additionalProperties"])

	// Test RateLimitConfig for double, int64, enum, map
	rl := &configv1.RateLimitConfig{}
	schemaMap = GenerateSchemaMapFromProto(rl.ProtoReflect())
	defs = schemaMap["$defs"].(map[string]interface{})
	rlDef := defs["mcpany.config.v1.RateLimitConfig"].(map[string]interface{})
	props := rlDef["properties"].(map[string]interface{})

	// double -> number
	rps := props["requests_per_second"].(map[string]interface{})
	assert.Equal(t, "number", rps["type"])

	// int64 -> string (ProtoJSON rule)
	burst := props["burst"].(map[string]interface{})
	assert.Equal(t, "string", burst["type"])

	// enum -> string
	storage := props["storage"].(map[string]interface{})
	assert.Equal(t, "string", storage["type"])

	// map -> object with additionalProperties
	toolLimits := props["tool_limits"].(map[string]interface{})
	assert.Equal(t, "object", toolLimits["type"])
}

func TestGenerateSchemaFromProto_Recursion(t *testing.T) {
	// RateLimitConfig contains a map to itself: map<string, RateLimitConfig> tool_limits
	rl := &configv1.RateLimitConfig{}
	schemaMap := GenerateSchemaMapFromProto(rl.ProtoReflect())
	defs := schemaMap["$defs"].(map[string]interface{})

	rlDef := defs["mcpany.config.v1.RateLimitConfig"].(map[string]interface{})
	props := rlDef["properties"].(map[string]interface{})

	// Check that tool_limits refers back to RateLimitConfig
	toolLimits := props["tool_limits"].(map[string]interface{})
	additionalProps := toolLimits["additionalProperties"].(map[string]interface{})
	assert.Equal(t, "#/$defs/mcpany.config.v1.RateLimitConfig", additionalProps["$ref"])
}
