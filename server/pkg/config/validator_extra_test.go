// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestValidate_ExtraServices(t *testing.T) {
	tests := []struct {
		name                string
		config              *configv1.McpAnyServerConfig
		expectedErrorCount  int
		expectedErrorString string
	}{
		{
			name: "valid graphql service",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				// Use struct construction
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("graphql-valid"),
						ServiceConfig: &configv1.UpstreamServiceConfig_GraphqlService{
							GraphqlService: &configv1.GraphQLUpstreamService{
								Address: proto.String("http://example.com/graphql"),
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid graphql service - bad scheme",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "graphql-bad-scheme",
							"graphql_service": {
								"address": "ftp://example.com/graphql"
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "graphql-bad-scheme": invalid graphql target_address scheme: ftp`,
		},
		{
			name: "valid webrtc service",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				// Use struct construction
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("webrtc-valid"),
						ServiceConfig: &configv1.UpstreamServiceConfig_WebrtcService{
							WebrtcService: &configv1.WebrtcUpstreamService{
								Address: proto.String("http://example.com/webrtc"),
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid webrtc service - bad scheme",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "webrtc-bad-scheme",
							"webrtc_service": {
								"address": "ftp://example.com/webrtc"
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "webrtc-bad-scheme": invalid webrtc target_address scheme: ftp`,
		},
		{
			name: "valid upstream service collection",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_service_collections": [
						{
							"name": "collection-1",
							"http_url": "http://example.com/collection"
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid upstream service collection - empty name",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_service_collections": [
						{
							"name": "",
							"http_url": "http://example.com/collection"
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "": collection name is empty`,
		},
		{
			name: "invalid upstream service collection - empty url",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_service_collections": [
						{
							"name": "collection-no-url",
							"http_url": ""
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "collection-no-url": collection http_url is empty`,
		},
		{
			name: "invalid upstream service collection - bad url scheme",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_service_collections": [
						{
							"name": "collection-bad-scheme",
							"http_url": "ftp://example.com/collection"
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "collection-bad-scheme": invalid collection http_url scheme: ftp`,
		},
		{
			name: "valid upstream service collection with auth",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_service_collections": [
						{
							"name": "collection-auth",
							"http_url": "http://example.com/collection",
							"authentication": {
								"basic_auth": {
									"username": "user",
									"password": { "plainText": "pass" }
								}
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "duplicate service name",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "service-1",
							"http_service": { "address": "http://example.com/1" }
						},
						{
							"name": "service-1",
							"http_service": { "address": "http://example.com/2" }
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "service-1": duplicate service name found`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := Validate(context.Background(), tt.config, Server)
			if tt.expectedErrorCount > 0 {
				require.NotEmpty(t, validationErrors, "expected validation errors but got none")
				found := false
				for _, err := range validationErrors {
					if err.Error() == tt.expectedErrorString {
						found = true
						break
					}
				}
				if !found {
					assert.EqualError(t, &validationErrors[0], tt.expectedErrorString)
				}
			} else {
				assert.Empty(t, validationErrors)
			}
		})
	}
}
