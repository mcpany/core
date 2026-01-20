package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func httpMethodPtr(m configv1.HttpCallDefinition_HttpMethod) *configv1.HttpCallDefinition_HttpMethod {
	return &m
}

func TestValidateToolsReferenceCalls(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                string
		serviceConfig       *configv1.UpstreamServiceConfig
		expectedErrorCount  int
		expectedErrorString string
	}{
		{
			name: "Valid HTTP Tool Call Reference",
			serviceConfig: &configv1.UpstreamServiceConfig{
				Name: proto.String("valid-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String("http://example.com"),
						Tools: []*configv1.ToolDefinition{
							{
								Name:   proto.String("get_weather"),
								CallId: proto.String("get_weather_call"),
							},
						},
						Calls: map[string]*configv1.HttpCallDefinition{
							"get_weather_call": {
								EndpointPath: proto.String("/weather"),
								Method:       httpMethodPtr(configv1.HttpCallDefinition_HTTP_METHOD_GET),
							},
						},
					},
				},
			},
			expectedErrorCount: 0,
		},
		{
			name: "Invalid HTTP Tool Call Reference (Dangling)",
			serviceConfig: &configv1.UpstreamServiceConfig{
				Name: proto.String("invalid-http"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String("http://example.com"),
						Tools: []*configv1.ToolDefinition{
							{
								Name:   proto.String("get_weather"),
								CallId: proto.String("missing_call_id"),
							},
						},
						Calls: map[string]*configv1.HttpCallDefinition{
							"other_call": {
								EndpointPath: proto.String("/weather"),
							},
						},
					},
				},
			},
			expectedErrorCount:  1,
			expectedErrorString: `tool "get_weather" references non-existent call_id "missing_call_id"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrError(ctx, tt.serviceConfig)
			if tt.expectedErrorCount > 0 {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectedErrorString)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
