// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestConvertToolDefinitionToProto_Coverage(t *testing.T) {
	// 1. Basic Tool Definition
	def := configv1.ToolDefinition_builder{
		Name:        proto.String("test-tool"),
		Description: proto.String("A test tool"),
		ServiceId:   proto.String("test-service"),
	}.Build()

	toolProto, err := ConvertToolDefinitionToProto(def, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "test-tool", toolProto.GetName())
	assert.Equal(t, "A test tool", toolProto.GetDescription())
	assert.Equal(t, "test-service", toolProto.GetServiceId())

	// 2. InputSchema via Map (JSON object)
	schemaMap := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"arg": map[string]interface{}{"type": "string"},
		},
	}
	// Use structpb for InputSchema directly if possible, or pass it via ConvertToolDefinitionToProto args?
	// ConvertToolDefinitionToProto signature: (toolDef, inputSchema, outputSchema)
	// In the test I was trying to set InputSchema in ToolDefinition, but ToolDefinition doesn't seem to have InputSchema field directly?
	// Wait, ToolDefinition in tool.proto likely has InputSchema if I checked it?
	// But ConvertToolDefinitionToProto takes extra args.
	// Let's check the signature in converters.go line 130:
	// func ConvertToolDefinitionToProto(toolDef *configv1.ToolDefinition, inputSchema, outputSchema *structpb.Struct) ...

	inputStruct, err := structpb.NewStruct(schemaMap)
	require.NoError(t, err)

	toolProto2, err := ConvertToolDefinitionToProto(def, inputStruct, nil)
	require.NoError(t, err)
	assert.NotNil(t, toolProto2.GetAnnotations().GetInputSchema())
	assert.Equal(t, "object", toolProto2.GetAnnotations().GetInputSchema().GetFields()["type"].GetStringValue())
}

func TestConvertToolDefinitionToProto_ErrorCases(t *testing.T) {
	// Nil definition
	_, err := ConvertToolDefinitionToProto(nil, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot convert nil tool definition")
}
