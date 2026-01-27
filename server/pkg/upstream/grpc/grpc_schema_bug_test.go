// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGRPCUpstream_SchemaBug_ExplicitConfig(t *testing.T) {
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	// Use config-based tool definition (not reflection auto-discovery)
	grpcService := configv1.GrpcUpstreamService_builder{
		Address:      proto.String(addr),
		UseReflection: proto.Bool(true), // Reflection used to fetch descriptors, but we define tools explicitly
		Tools: []*configv1.ToolDefinition{
			configv1.ToolDefinition_builder{
				Name:   proto.String("GetWeatherConfig"),
				CallId: proto.String("weather_call"),
			}.Build(),
		},
		Calls: map[string]*configv1.GrpcCallDefinition{
			"weather_call": configv1.GrpcCallDefinition_builder{
				Id:      proto.String("weather_call"),
				Service: proto.String("examples.weather.v1.WeatherService"),
				Method:  proto.String("GetWeather"),
			}.Build(),
		},
	}.Build()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String("weather-service-bug"),
		GrpcService: grpcService,
	}.Build()

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	// Verify the tool's schema
	toolName := serviceID + ".GetWeatherConfig"
	tool, ok := tm.GetTool(toolName)
	require.True(t, ok, "Tool should be registered")

	inputSchema := tool.Tool().GetAnnotations().GetInputSchema()
	require.NotNil(t, inputSchema)

	// Debug logging
	t.Logf("Input Schema Keys: %v", inputSchema.GetFields())

	// Check if "type" field exists and is "object"
	typeField := inputSchema.GetFields()["type"]
	require.NotNil(t, typeField, "InputSchema should have 'type' field")
	assert.Equal(t, "object", typeField.GetStringValue())

	// Check if "properties" field exists
	propsField := inputSchema.GetFields()["properties"]
	require.NotNil(t, propsField, "InputSchema should have 'properties' field")
}

func TestGRPCUpstream_SchemaOverride(t *testing.T) {
	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)
	tm := NewMockToolManager()

	// Define explicit input schema
	customSchema, err := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"custom_field": map[string]any{
				"type":        "string",
				"description": "A custom field overriding auto-generation",
			},
		},
		"required": []any{"custom_field"},
	})
	require.NoError(t, err)

	grpcService := configv1.GrpcUpstreamService_builder{
		Address:       proto.String(addr),
		UseReflection: proto.Bool(true),
		Tools: []*configv1.ToolDefinition{
			configv1.ToolDefinition_builder{
				Name:   proto.String("GetWeatherWithOverride"),
				CallId: proto.String("weather_call_override"),
			}.Build(),
		},
		Calls: map[string]*configv1.GrpcCallDefinition{
			"weather_call_override": configv1.GrpcCallDefinition_builder{
				Id:          proto.String("weather_call_override"),
				Service:     proto.String("examples.weather.v1.WeatherService"),
				Method:      proto.String("GetWeather"),
				InputSchema: customSchema,
			}.Build(),
		},
	}.Build()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String("weather-service-override"),
		GrpcService: grpcService,
	}.Build()

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)

	toolName := serviceID + ".GetWeatherWithOverride"
	registeredTool, ok := tm.GetTool(toolName)
	require.True(t, ok, "Tool should be registered")

	inputSchema := registeredTool.Tool().GetAnnotations().GetInputSchema()
	require.NotNil(t, inputSchema)

	// Debug logging
	t.Logf("Input Schema: %v", inputSchema)

	// Check for custom_field
	props := inputSchema.GetFields()["properties"].GetStructValue().GetFields()
	_, found := props["custom_field"]
	assert.True(t, found, "Input schema should contain 'custom_field' from explicit config")

	// Check that required fields list matches explicit config
	required := inputSchema.GetFields()["required"].GetListValue().GetValues()
	assert.NotEmpty(t, required)
	assert.Equal(t, "custom_field", required[0].GetStringValue())
}
