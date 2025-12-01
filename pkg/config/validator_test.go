/*
 * Copyright 2025 Author(s) of MCP Any
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
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestValidate(t *testing.T) {
	// Create temporary files for mTLS tests
	clientCertFile, err := os.CreateTemp("", "client.crt")
	require.NoError(t, err)
	defer os.Remove(clientCertFile.Name())

	clientKeyFile, err := os.CreateTemp("", "client.key")
	require.NoError(t, err)
	defer os.Remove(clientKeyFile.Name())

	caCertFile, err := os.CreateTemp("", "ca.crt")
	require.NoError(t, err)
	defer os.Remove(caCertFile.Name())

	tests := []struct {
		name                string
		config              *configv1.McpAnyServerConfig
		expectedErrorCount  int
		expectedErrorString string
	}{
		{
			name: "valid grpc service",
			config: (&configv1.McpAnyServerConfig_builder{
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
			expectedErrorCount: 0,
		},
		{
			name: "invalid websocket service - invalid scheme",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("ws-svc-1"),
						WebsocketService: (&configv1.WebsocketUpstreamService_builder{
							Address: proto.String("http://localhost:8080"),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "ws-svc-1": invalid websocket target_address scheme: http`,
		},
		{
			name: "valid websocket service",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("ws-svc-1"),
						WebsocketService: (&configv1.WebsocketUpstreamService_builder{
							Address: proto.String("ws://localhost:8080"),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid http service - empty address",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("http-svc-1"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String(""),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "http-svc-1": http service has empty target_address`,
		},
		{
			name: "invalid http service - invalid url",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("http-svc-2"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("not a url"),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "http-svc-2": invalid http target_address: not a url`,
		},
		{
			name: "invalid http service - invalid scheme",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("http-svc-3"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("ftp://localhost:8080"),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "http-svc-3": invalid http target_address scheme: ftp`,
		},
		{
			name: "valid http service",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("http-svc-1"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("http://localhost:8080"),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount: 0,
		},
		{
			name: "valid openapi service",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("openapi-svc-1"),
						OpenapiService: (&configv1.OpenapiUpstreamService_builder{
							Address: proto.String("http://localhost:8080"),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount: 0,
		},
		{
			name: "valid mcp service (http)",
			config: (&configv1.McpAnyServerConfig_builder{
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
			expectedErrorCount: 0,
		},
		{
			name: "valid mcp service (stdio)",
			config: (&configv1.McpAnyServerConfig_builder{
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
			expectedErrorCount: 0,
		},
		{
			name: "invalid grpc service - empty address",
			config: (&configv1.McpAnyServerConfig_builder{
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
			expectedErrorCount:  1,
			expectedErrorString: `service "grpc-svc-1": gRPC service has empty target_address`,
		},
		{
			name: "invalid openapi service - invalid address should be an error",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("openapi-svc-1"),
						OpenapiService: (&configv1.OpenapiUpstreamService_builder{
							Address: proto.String("not a url"),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "openapi-svc-1": invalid openapi target_address: not a url`,
		},
		{
			name: "invalid openapi service - no address or spec path",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("openapi-svc-2"),
						OpenapiService: (&configv1.OpenapiUpstreamService_builder{}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "openapi-svc-2": openapi service must have either an address or a spec path`,
		},
		{
			name: "invalid mcp service - no connection",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name:       proto.String("mcp-svc-1"),
						McpService: (&configv1.McpUpstreamService_builder{}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "mcp-svc-1": mcp service has no connection_type`,
		},
		{
			name: "invalid mcp service - empty stdio command",
			config: (&configv1.McpAnyServerConfig_builder{
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
			expectedErrorCount:  1,
			expectedErrorString: `service "mcp-svc-2": mcp service with stdio_connection has empty command`,
		},
		{
			name: "duplicate service name",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("dup-svc"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("http://localhost:8080"),
						}).Build(),
					}).Build(),
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("dup-svc"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("http://localhost:8081"),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "dup-svc": duplicate service name found`,
		},
		{
			name: "duplicate service name and another error",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("dup-svc"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("http://localhost:8080"),
						}).Build(),
					}).Build(),
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("dup-svc"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String(""),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount:  2,
			expectedErrorString: `service "dup-svc": duplicate service name found`,
		},
		{
			name: "invalid basic auth - empty password",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("basic-auth-svc-1"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("http://localhost:8080"),
						}).Build(),
						UpstreamAuthentication: (&configv1.UpstreamAuthentication_builder{
							BasicAuth: (&configv1.UpstreamBasicAuth_builder{
								Username: proto.String("user"),
								Password: (&configv1.SecretValue_builder{
									PlainText: proto.String(""),
								}).Build(),
							}).Build(),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "basic-auth-svc-1": basic auth 'password' is empty`,
		},
		{
			name: "invalid api key auth - empty header name",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("api-key-svc-1"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("http://localhost:8080"),
						}).Build(),
						UpstreamAuthentication: (&configv1.UpstreamAuthentication_builder{
							ApiKey: (&configv1.UpstreamAPIKeyAuth_builder{
								HeaderName: proto.String(""),
								ApiKey: (&configv1.SecretValue_builder{
									PlainText: proto.String("some-key"),
								}).Build(),
							}).Build(),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "api-key-svc-1": api key 'header_name' is empty`,
		},
		{
			name: "invalid basic auth - empty username",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("basic-auth-svc-2"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("http://localhost:8080"),
						}).Build(),
						UpstreamAuthentication: (&configv1.UpstreamAuthentication_builder{
							BasicAuth: (&configv1.UpstreamBasicAuth_builder{
								Username: proto.String(""),
								Password: (&configv1.SecretValue_builder{
									PlainText: proto.String("some-password"),
								}).Build(),
							}).Build(),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "basic-auth-svc-2": basic auth 'username' is empty`,
		},
		{
			name: "valid mtls auth - with ca cert",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mtls-svc-1"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("http://localhost:8080"),
						}).Build(),
						UpstreamAuthentication: (&configv1.UpstreamAuthentication_builder{
							Mtls: (&configv1.UpstreamMTLSAuth_builder{
								ClientCertPath: proto.String(clientCertFile.Name()),
								ClientKeyPath:  proto.String(clientKeyFile.Name()),
								CaCertPath:     proto.String(caCertFile.Name()),
							}).Build(),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount: 0,
		},
		{
			name: "valid mtls auth - without ca cert",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mtls-svc-2"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("http://localhost:8080"),
						}).Build(),
						UpstreamAuthentication: (&configv1.UpstreamAuthentication_builder{
							Mtls: (&configv1.UpstreamMTLSAuth_builder{
								ClientCertPath: proto.String(clientCertFile.Name()),
								ClientKeyPath:  proto.String(clientKeyFile.Name()),
							}).Build(),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid mtls auth - ca cert path not found",
			config: (&configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					(&configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mtls-svc-3"),
						HttpService: (&configv1.HttpUpstreamService_builder{
							Address: proto.String("http://localhost:8080"),
						}).Build(),
						UpstreamAuthentication: (&configv1.UpstreamAuthentication_builder{
							Mtls: (&configv1.UpstreamMTLSAuth_builder{
								ClientCertPath: proto.String(clientCertFile.Name()),
								ClientKeyPath:  proto.String(clientKeyFile.Name()),
								CaCertPath:     proto.String("nonexistent.crt"),
							}).Build(),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "mtls-svc-3": mtls 'ca_cert_path' not found: stat nonexistent.crt: no such file or directory`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := Validate(tt.config, Server)
			assert.Len(t, validationErrors, tt.expectedErrorCount)
			if tt.expectedErrorCount > 0 {
				require.NotEmpty(t, validationErrors)
				assert.EqualError(t, &validationErrors[0], tt.expectedErrorString)
			}
		})
	}
}

func TestValidateGlobalSettings_Validation(t *testing.T) {
	tests := []struct {
		name          string
		globalConfig  *configv1.GlobalSettings
		binaryType    BinaryType
		expectedError bool
	}{
		{
			name: "valid global settings",
			globalConfig: (&configv1.GlobalSettings_builder{
				McpListenAddress: proto.String("localhost:8081"),
			}).Build(),
			binaryType:    Server,
			expectedError: false,
		},
		{
			name: "empty bind address for server",
			globalConfig: (&configv1.GlobalSettings_builder{
				McpListenAddress: proto.String(""),
			}).Build(),
			binaryType:    Server,
			expectedError: false,
		},
		{
			name: "invalid bind address for server",
			globalConfig: (&configv1.GlobalSettings_builder{
				McpListenAddress: proto.String("invalid"),
			}).Build(),
			binaryType:    Server,
			expectedError: true,
		},
		{
			name: "empty bind address for worker",
			globalConfig: (&configv1.GlobalSettings_builder{
				McpListenAddress: proto.String(""),
			}).Build(),
			binaryType:    Worker,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGlobalSettings(tt.globalConfig, tt.binaryType)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOrError(t *testing.T) {
	tests := []struct {
		name        string
		service     *configv1.UpstreamServiceConfig
		expectError bool
	}{
		{
			name: "valid http service",
			service: (&configv1.UpstreamServiceConfig_builder{
				Name: proto.String("http-svc-1"),
				HttpService: (&configv1.HttpUpstreamService_builder{
					Address: proto.String("http://localhost:8080"),
				}).Build(),
			}).Build(),
			expectError: false,
		},
		{
			name: "invalid http service - empty address",
			service: (&configv1.UpstreamServiceConfig_builder{
				Name: proto.String("http-svc-1"),
				HttpService: (&configv1.HttpUpstreamService_builder{
					Address: proto.String(""),
				}).Build(),
			}).Build(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrError(tt.service)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
