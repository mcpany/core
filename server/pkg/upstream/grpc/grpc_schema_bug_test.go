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
	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(true) // Reflection used to fetch descriptors, but we define tools explicitly

	// Explicitly define tool and call
	grpcService.Tools = []*configv1.ToolDefinition{
		{
			Name:   proto.String("GetWeatherConfig"),
			CallId: proto.String("weather_call"),
		},
	}
	grpcService.Calls = map[string]*configv1.GrpcCallDefinition{
		"weather_call": {
			Id:      proto.String("weather_call"),
			Service: proto.String("examples.weather.v1.WeatherService"),
			Method:  proto.String("GetWeather"),
		},
	}

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("weather-service-bug")
	serviceConfig.SetGrpcService(grpcService)

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
