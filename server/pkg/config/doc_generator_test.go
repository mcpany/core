// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGenerateDocumentation(t *testing.T) {
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type": "string",
			},
		},
	})

	cfg := func() *configv1.McpAnyServerConfig {
		c := &configv1.McpAnyServerConfig{}
		svc := configv1.UpstreamServiceConfig_builder{}.Build()
		svc.SetName("weather")

		httpSvc := configv1.HttpUpstreamService_builder{}.Build()
		httpSvc.SetAddress("http://example.com")

		tool := configv1.ToolDefinition_builder{}.Build()
		tool.SetName("get_weather")
		tool.SetDescription("Get the weather")
		tool.SetInputSchema(inputSchema)
		tool.SetServiceId("weather-id")
		tool.SetCallId("weather_call")
		httpSvc.SetTools([]*configv1.ToolDefinition{tool})
		httpSvc.SetCalls(map[string]*configv1.HttpCallDefinition{
			"weather_call": configv1.HttpCallDefinition_builder{
				Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
				EndpointPath: proto.String("/weather"),
				InputSchema:  inputSchema,
			}.Build(),
		})

		svc.SetHttpService(httpSvc)
		c.SetUpstreamServices([]*configv1.UpstreamServiceConfig{svc})
		return c
	}()

	doc, err := config.GenerateDocumentation(context.Background(), cfg)
	assert.NoError(t, err)
	assert.Contains(t, doc, "# Available Tools")
	assert.Contains(t, doc, "## `weather.get_weather`")
	assert.Contains(t, doc, "Get the weather")
	assert.Contains(t, doc, "\"query\"")
}
