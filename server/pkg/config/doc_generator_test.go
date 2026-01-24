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

func newHttpDocParam(name string, typ configv1.ParameterType) *configv1.HttpParameterMapping {
	p := &configv1.HttpParameterMapping{}
	s := &configv1.ParameterSchema{}
	s.SetName(name)
	s.SetType(typ)
	p.SetSchema(s)
	return p
}

func newHttpDocCall(path string, method configv1.HttpCallDefinition_HttpMethod, params []*configv1.HttpParameterMapping) *configv1.HttpCallDefinition {
	c := &configv1.HttpCallDefinition{}
	c.SetEndpointPath(path)
	c.SetMethod(method)
	c.SetParameters(params)
	return c
}

func newToolDef(name, desc, svcId, callId string, input *structpb.Struct) *configv1.ToolDefinition {
	t := &configv1.ToolDefinition{}
	t.SetName(name)
	t.SetDescription(desc)
	t.SetServiceId(svcId)
	t.SetCallId(callId)
	t.SetInputSchema(input)
	return t
}

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
							"weather_call": newHttpDocCall(
								"/weather",
								configv1.HttpCallDefinition_HTTP_METHOD_GET,
								[]*configv1.HttpParameterMapping{
									newHttpDocParam("query", configv1.ParameterType_STRING),
								},
							),
						},
						Tools: []*configv1.ToolDefinition{
							newToolDef("get_weather", "Get the weather", "weather-id", "weather_call", inputSchema),
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
