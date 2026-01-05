// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	configv1 "github.com/mcpany/core/proto/config/v1"
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

	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("weather"),
				Id:   proto.String("weather-id"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String("http://example.com"),
						Calls: map[string]*configv1.HttpCallDefinition{
							"weather_call": {
								EndpointPath: proto.String("/weather"),
								Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
								Parameters: []*configv1.HttpParameterMapping{
									{
										Schema: &configv1.ParameterSchema{
											Name: proto.String("query"),
											Type: configv1.ParameterType_STRING.Enum(),
										},
									},
								},
							},
						},
						Tools: []*configv1.ToolDefinition{
							{
								Name:        proto.String("get_weather"),
								Description: proto.String("Get the weather"),
								InputSchema: inputSchema,
								ServiceId:   proto.String("weather-id"),
								CallId:      proto.String("weather_call"),
							},
						},
					},
				},
			},
		},
	}

	doc, err := config.GenerateDocumentation(context.Background(), cfg)
	assert.NoError(t, err)
	assert.Contains(t, doc, "# Available Tools")
	assert.Contains(t, doc, "## `weather.get_weather`")
	assert.Contains(t, doc, "Get the weather")
	assert.Contains(t, doc, "\"query\"")
}
