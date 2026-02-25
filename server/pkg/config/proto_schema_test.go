// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resolveRef(t *testing.T, rootSchema map[string]interface{}, obj map[string]interface{}) map[string]interface{} {
	ref, ok := obj["$ref"].(string)
	if !ok {
		return obj
	}
	defs, ok := rootSchema["$defs"].(map[string]interface{})
	require.True(t, ok, "$defs not found in root schema")

	name := strings.TrimPrefix(ref, "#/$defs/")
	def, ok := defs[name]
	require.True(t, ok, "Definition %s not found", name)

	return def.(map[string]interface{})
}

func TestGenerateSchemaMapFromProto_SchemaVersion(t *testing.T) {
	msg := &configv1.CacheConfig{}
	schema := GenerateSchemaMapFromProto(msg.ProtoReflect())
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", schema["$schema"])
}

func TestGenerateSchemaMapFromProto_PrimitiveTypes(t *testing.T) {
	// CacheConfig uses bool, Duration (string pattern), string, float
	msg := &configv1.CacheConfig{}
	schema := GenerateSchemaMapFromProto(msg.ProtoReflect())

	rootDef := resolveRef(t, schema, schema)
	assert.Equal(t, "object", rootDef["type"])
	props := rootDef["properties"].(map[string]interface{})

	// bool
	assert.Equal(t, map[string]interface{}{"type": "boolean"}, props["is_enabled"])

	// string
	assert.Equal(t, map[string]interface{}{"type": "string"}, props["strategy"])

	// Duration
	ttlSchemaRef := props["ttl"].(map[string]interface{})
	ttlSchema := resolveRef(t, schema, ttlSchemaRef)

	assert.Equal(t, "string", ttlSchema["type"])
	assert.Equal(t, "^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$", ttlSchema["pattern"])

	// Check additionalProperties: false at root
	assert.Equal(t, false, rootDef["additionalProperties"])
}

func TestGenerateSchemaMapFromProto_Enums(t *testing.T) {
	msg := &configv1.OutputTransformer{}
	schema := GenerateSchemaMapFromProto(msg.ProtoReflect())

	rootDef := resolveRef(t, schema, schema)
	props := rootDef["properties"].(map[string]interface{})

	// Enum should be string
	// The implementation simplifies enum to just "string"
	assert.Equal(t, map[string]interface{}{"type": "string"}, props["format"])
}

func TestGenerateSchemaMapFromProto_Maps(t *testing.T) {
	msg := &configv1.OutputTransformer{}
	schema := GenerateSchemaMapFromProto(msg.ProtoReflect())

	rootDef := resolveRef(t, schema, schema)
	props := rootDef["properties"].(map[string]interface{})

	// map<string, string> extraction_rules
	rules := props["extraction_rules"].(map[string]interface{})
	assert.Equal(t, "object", rules["type"])
	assert.Equal(t, map[string]interface{}{"type": "string"}, rules["additionalProperties"])
}

func TestGenerateSchemaMapFromProto_Struct(t *testing.T) {
	msg := &configv1.HttpCallDefinition{}
	schema := GenerateSchemaMapFromProto(msg.ProtoReflect())

	rootDef := resolveRef(t, schema, schema)
	props := rootDef["properties"].(map[string]interface{})

	// google.protobuf.Struct input_schema
	inputSchemaRef := props["input_schema"].(map[string]interface{})
	inputSchema := resolveRef(t, schema, inputSchemaRef)

	assert.Equal(t, "object", inputSchema["type"])
	assert.Equal(t, true, inputSchema["additionalProperties"])
}

func TestGenerateSchemaMapFromProto_Repeated(t *testing.T) {
	msg := &configv1.HttpCallDefinition{}
	schema := GenerateSchemaMapFromProto(msg.ProtoReflect())

	rootDef := resolveRef(t, schema, schema)
	props := rootDef["properties"].(map[string]interface{})

	// repeated HttpParameterMapping parameters
	params := props["parameters"].(map[string]interface{})
	assert.Equal(t, "array", params["type"])

	items := params["items"].(map[string]interface{})
	// Items should be a ref to HttpParameterMapping
	assert.Contains(t, items, "$ref")
	assert.Equal(t, "#/$defs/mcpany.config.v1.HttpParameterMapping", items["$ref"])

	// Verify definitions exist
	defs := schema["$defs"].(map[string]interface{})
	assert.Contains(t, defs, "mcpany.config.v1.HttpParameterMapping")
}

func TestGenerateSchemaMapFromProto_NestedAndRef(t *testing.T) {
	msg := &configv1.CacheConfig{}
	schema := GenerateSchemaMapFromProto(msg.ProtoReflect())

	rootDef := resolveRef(t, schema, schema)
	props := rootDef["properties"].(map[string]interface{})

	// SemanticCacheConfig semantic_config
	semantic := props["semantic_config"].(map[string]interface{})
	assert.Contains(t, semantic, "$ref")

	defs := schema["$defs"].(map[string]interface{})
	semanticDef := defs["mcpany.config.v1.SemanticCacheConfig"].(map[string]interface{})
	semanticProps := semanticDef["properties"].(map[string]interface{})

	// float
	assert.Equal(t, map[string]interface{}{"type": "number"}, semanticProps["similarity_threshold"])

	// OneOf (provider_config) -> flattened fields in properties
	assert.Contains(t, semanticProps, "openai")
	assert.Contains(t, semanticProps, "ollama")
	assert.Contains(t, semanticProps, "http")
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
	// This triggers AddResource parsing failure or compilation failure
	badSchema := map[string]interface{}{
		"$id": 123,
	}

	schema, err := CompileSchema(badSchema)
	assert.Error(t, err)
	assert.Nil(t, schema)
}

func TestCompileSchema_Success(t *testing.T) {
	msg := &configv1.CacheConfig{}
	schemaMap := GenerateSchemaMapFromProto(msg.ProtoReflect())

	schema, err := CompileSchema(schemaMap)
	require.NoError(t, err)
	assert.NotNil(t, schema)

	// Validate a valid JSON against the schema
	validJSON := `{
		"is_enabled": true,
		"ttl": "10s",
		"strategy": "lru"
	}`
	var v interface{}
	require.NoError(t, json.Unmarshal([]byte(validJSON), &v))
	err = schema.Validate(v)
	assert.NoError(t, err)

	// Validate invalid JSON (wrong type)
	invalidJSON := `{
		"is_enabled": "yes"
	}`
	require.NoError(t, json.Unmarshal([]byte(invalidJSON), &v))
	err = schema.Validate(v)
	assert.Error(t, err)
}
