package config_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/stretchr/testify/assert"
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
		svc := &configv1.UpstreamServiceConfig{}
		svc.SetName("weather")
		svc.SetId("weather-id")

		httpSvc := &configv1.HttpUpstreamService{}
		httpSvc.SetAddress("http://example.com")

		callDef := &configv1.HttpCallDefinition{}
		callDef.SetEndpointPath("/weather")
		callDef.SetMethod(configv1.HttpCallDefinition_HTTP_METHOD_GET)

		param := &configv1.HttpParameterMapping{}
		schema := &configv1.ParameterSchema{}
		schema.SetName("query")
		schema.SetType(configv1.ParameterType_STRING)
		param.SetSchema(schema)
		callDef.SetParameters([]*configv1.HttpParameterMapping{param})

		httpSvc.SetCalls(map[string]*configv1.HttpCallDefinition{
			"weather_call": callDef,
		})

		tool := &configv1.ToolDefinition{}
		tool.SetName("get_weather")
		tool.SetDescription("Get the weather")
		tool.SetInputSchema(inputSchema)
		tool.SetServiceId("weather-id")
		tool.SetCallId("weather_call")
		httpSvc.SetTools([]*configv1.ToolDefinition{tool})

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
