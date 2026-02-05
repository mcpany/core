package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestValidateSQLService_Coverage(t *testing.T) {
	tests := []struct {
		name                string
		config              *configv1.McpAnyServerConfig
		expectedErrorString string
		shouldFail          bool
	}{
		{
			name: "Valid SQL Service",
			config: func() *configv1.McpAnyServerConfig {
				sqlSvc := configv1.SqlUpstreamService_builder{
					Driver: proto.String("postgres"),
					Dsn:    proto.String("postgres://user:pass@127.0.0.1:5432/db"),
					Calls: map[string]*configv1.SqlCallDefinition{
						"my-query": configv1.SqlCallDefinition_builder{
							Query: proto.String("SELECT 1"),
							InputSchema: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"type": {Kind: &structpb.Value_StringValue{StringValue: "object"}},
								},
							},
							OutputSchema: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"type": {Kind: &structpb.Value_StringValue{StringValue: "array"}},
								},
							},
						}.Build(),
					},
				}.Build()

				svcConfig := configv1.UpstreamServiceConfig_builder{
					Name:       proto.String("sql-valid"),
					SqlService: sqlSvc,
				}.Build()

				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svcConfig},
				}.Build()
			}(),
			shouldFail: false,
		},
		{
			name: "Invalid Output Schema",
			config: func() *configv1.McpAnyServerConfig {
				sqlSvc := configv1.SqlUpstreamService_builder{
					Driver: proto.String("postgres"),
					Dsn:    proto.String("postgres://user:pass@127.0.0.1:5432/db"),
					Calls: map[string]*configv1.SqlCallDefinition{
						"my-query": configv1.SqlCallDefinition_builder{
							Query: proto.String("SELECT 1"),
							OutputSchema: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"type": {Kind: &structpb.Value_NumberValue{NumberValue: 1}},
								},
							},
						}.Build(),
					},
				}.Build()

				svcConfig := configv1.UpstreamServiceConfig_builder{
					Name:       proto.String("sql-invalid-output"),
					SqlService: sqlSvc,
				}.Build()

				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svcConfig},
				}.Build()
			}(),
			shouldFail:          true,
			expectedErrorString: `service "sql-invalid-output": sql call "my-query" output_schema error: schema 'type' must be a string`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := Validate(context.Background(), tt.config, Server)
			if tt.shouldFail {
				if len(validationErrors) == 0 {
					t.Fatalf("Expected validation error %q, but got none.", tt.expectedErrorString)
				}
				assert.EqualError(t, &validationErrors[0], tt.expectedErrorString)
			} else {
				assert.Empty(t, validationErrors)
			}
		})
	}
}
