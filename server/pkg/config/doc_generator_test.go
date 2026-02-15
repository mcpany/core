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
		param := configv1.HttpParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name: proto.String("query"),
				Type: configv1.ParameterType_STRING.Enum(),
			}.Build(),
		}.Build()

		callDef := configv1.HttpCallDefinition_builder{
			EndpointPath: proto.String("/weather"),
			Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
			Parameters:   []*configv1.HttpParameterMapping{param},
		}.Build()

		tool := configv1.ToolDefinition_builder{
			Name:        proto.String("get_weather"),
			Description: proto.String("Get the weather"),
			InputSchema: inputSchema,
			ServiceId:   proto.String("weather-id"),
			CallId:      proto.String("weather_call"),
		}.Build()

		httpSvc := configv1.HttpUpstreamService_builder{
			Address: proto.String("http://example.com"),
			Calls: map[string]*configv1.HttpCallDefinition{
				"weather_call": callDef,
			},
			Tools: []*configv1.ToolDefinition{tool},
		}.Build()

		svc := configv1.UpstreamServiceConfig_builder{
			Name:        proto.String("weather"),
			Id:          proto.String("weather-id"),
			HttpService: httpSvc,
		}.Build()

		return configv1.McpAnyServerConfig_builder{
			UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
		}.Build()
	}()

	doc, err := config.GenerateDocumentation(context.Background(), cfg)
	assert.NoError(t, err)
	assert.Contains(t, doc, "# Available Tools")
	assert.Contains(t, doc, "## `weather.get_weather`")
	assert.Contains(t, doc, "Get the weather")
	assert.Contains(t, doc, "\"query\"")
}
