/*
 * Copyright 2025 Author(s) of MCPXY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"testing"

	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name          string
		config        *configv1.McpxServerConfig
		expectedCount int
		expectError   bool
	}{
		{
			name: "valid grpc service",
			config: (&configv1.McpxServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("grpc-svc-1"),
						GrpcService: (&configv1.GrpcUpstreamService_builder{
							Address:       proto.String("grpc://localhost:50051"),
							UseReflection: proto.Bool(true),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "invalid http service - empty address",
			config: (&configv1.McpxServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("http-svc-1"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String(""),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "invalid http service - invalid url",
			config: (&configv1.McpxServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("http-svc-2"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("not a url"),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "valid http service",
			config: (&configv1.McpxServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("http-svc-1"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("http://localhost:8080"),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "valid openapi service",
			config: (&configv1.McpxServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("openapi-svc-1"),
						OpenapiService: (&configv1.OpenapiUpstreamService_builder{
							Address: proto.String("http://localhost:8080"),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "valid mcp service (http)",
			config: (&configv1.McpxServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mcp-svc-1"),
						McpService: (&configv1.McpUpstreamService_builder{
							HttpConnection: (&configv1.McpStreamableHttpConnection_builder{
								HttpAddress: proto.String("http://localhost:8080"),
							}).Build(),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "valid mcp service (stdio)",
			config: (&configv1.McpxServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mcp-svc-2"),
						McpService: (&configv1.McpUpstreamService_builder{
							StdioConnection: (&configv1.McpStdioConnection_builder{
								Command: proto.String("echo"),
							}).Build(),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "invalid grpc service - empty address",
			config: (&configv1.McpxServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("grpc-svc-1"),
						GrpcService: (&configv1.GrpcUpstreamService_builder{
							Address:       proto.String(""),
							UseReflection: proto.Bool(true),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "invalid openapi service - invalid address",
			config: (&configv1.McpxServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("openapi-svc-1"),
						OpenapiService: (&configv1.OpenapiUpstreamService_builder{
							Address: proto.String("not a url"),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "invalid mcp service - no connection",
			config: (&configv1.McpxServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name:       proto.String("mcp-svc-1"),
						McpService: (&configv1.McpUpstreamService_builder{}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "invalid mcp service - empty stdio command",
			config: (&configv1.McpxServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mcp-svc-2"),
						McpService: (&configv1.McpUpstreamService_builder{
							StdioConnection: (&configv1.McpStdioConnection_builder{
								Command: proto.String(""),
							}).Build(),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validatedConfig, err := Validate(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, validatedConfig.GetUpstreamServices(), tt.expectedCount)
		})
	}
}
