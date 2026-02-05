package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestValidateSQLService_MissingValidation(t *testing.T) {
	tests := []struct {
		name                string
		config              *configv1.McpAnyServerConfig
		expectedErrorString string
		shouldFail          bool // If true, we expect validation to fail.
	}{
		{
			name: "invalid sql service - call with empty query",
			config: func() *configv1.McpAnyServerConfig {
				sqlSvc := configv1.SqlUpstreamService_builder{
					Driver: proto.String("postgres"),
					Dsn:    proto.String("postgres://user:pass@127.0.0.1:5432/db"),
					Calls: map[string]*configv1.SqlCallDefinition{
						"my-query": configv1.SqlCallDefinition_builder{
							Query: proto.String(""),
						}.Build(),
					},
				}.Build()

				svc := configv1.UpstreamServiceConfig_builder{
					Name:       proto.String("sql-empty-query"),
					SqlService: sqlSvc,
				}.Build()

				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
				}.Build()
			}(),
			expectedErrorString: `service "sql-empty-query": sql call "my-query" query is empty`,
			shouldFail:          true,
		},
		{
			name: "invalid sql service - call with invalid input schema",
			config: func() *configv1.McpAnyServerConfig {
				sqlSvc := configv1.SqlUpstreamService_builder{
					Driver: proto.String("postgres"),
					Dsn:    proto.String("postgres://user:pass@127.0.0.1:5432/db"),
					Calls: map[string]*configv1.SqlCallDefinition{
						"my-query": configv1.SqlCallDefinition_builder{
							Query: proto.String("SELECT * FROM users"),
							InputSchema: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}}, // Invalid type
								},
							},
						}.Build(),
					},
				}.Build()

				svc := configv1.UpstreamServiceConfig_builder{
					Name:       proto.String("sql-invalid-schema"),
					SqlService: sqlSvc,
				}.Build()

				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
				}.Build()
			}(),
			expectedErrorString: `service "sql-invalid-schema": sql call "my-query" input_schema error: schema 'type' must be a string`,
			shouldFail:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := Validate(context.Background(), tt.config, Server)
			if tt.shouldFail {
				if len(validationErrors) == 0 {
					t.Fatalf("Expected validation error %q, but got none. This confirms the bug.", tt.expectedErrorString)
				} else {
					assert.EqualError(t, &validationErrors[0], tt.expectedErrorString)
				}
			}
		})
	}
}
