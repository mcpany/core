// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
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
				c := &configv1.McpAnyServerConfig{}
				svc := &configv1.UpstreamServiceConfig{}
				svc.SetName("sql-empty-query")

				sqlSvc := &configv1.SqlUpstreamService{}
				sqlSvc.SetDriver("postgres")
				sqlSvc.SetDsn("postgres://user:pass@127.0.0.1:5432/db")
				sqlSvc.SetCalls(map[string]*configv1.SqlCallDefinition{
					"my-query": func() *configv1.SqlCallDefinition {
						callDef := &configv1.SqlCallDefinition{}
						callDef.SetQuery("") // Empty query should be invalid
						return callDef
					}(),
				})
				svc.SetSqlService(sqlSvc)

				c.SetUpstreamServices([]*configv1.UpstreamServiceConfig{svc})
				return c
			}(),
			expectedErrorString: `service "sql-empty-query": sql call "my-query" query is empty`,
			shouldFail:          true,
		},
		{
			name: "invalid sql service - call with invalid input schema",
			config: func() *configv1.McpAnyServerConfig {
				c := &configv1.McpAnyServerConfig{}
				svc := &configv1.UpstreamServiceConfig{}
				svc.SetName("sql-invalid-schema")

				sqlSvc := &configv1.SqlUpstreamService{}
				sqlSvc.SetDriver("postgres")
				sqlSvc.SetDsn("postgres://user:pass@127.0.0.1:5432/db")
				sqlSvc.SetCalls(map[string]*configv1.SqlCallDefinition{
					"my-query": func() *configv1.SqlCallDefinition {
						callDef := &configv1.SqlCallDefinition{}
						callDef.SetQuery("SELECT * FROM users")
						callDef.SetInputSchema(&structpb.Struct{
							Fields: map[string]*structpb.Value{
								"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}}, // Invalid type
							},
						})
						return callDef
					}(),
				})
				svc.SetSqlService(sqlSvc)

				c.SetUpstreamServices([]*configv1.UpstreamServiceConfig{svc})
				return c
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
