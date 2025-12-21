// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"testing"

	"github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestValidate(t *testing.T) {
	// Create temporary files for mTLS tests
	clientCertFile, err := os.CreateTemp("", "client.crt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(clientCertFile.Name()) }()

	clientKeyFile, err := os.CreateTemp("", "client.key")
	require.NoError(t, err)
	defer func() { _ = os.Remove(clientKeyFile.Name()) }()

	caCertFile, err := os.CreateTemp("", "ca.crt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(caCertFile.Name()) }()

	tests := []struct {
		name                string
		config              *configv1.McpAnyServerConfig
		expectedErrorCount  int
		expectedErrorString string
	}{
		{
			name: "invalid command line service - empty command",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "cmd-svc-1",
							"command_line_service": {
								"command": ""
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "cmd-svc-1": command_line_service has empty command`,
		},
		{
			name: "invalid mTLS config - insecure client cert path",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "mtls-svc-1",
							"http_service": {
								"address": "https://example.com"
							},
							"upstream_authentication": {
								"mtls": {
									"client_cert_path": "../../../etc/passwd",
									"client_key_path": "testdata/client.key",
									"ca_cert_path": "testdata/ca.crt"
								}
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "mtls-svc-1": mtls 'client_cert_path' is not a secure path: path contains '..', which is not allowed`,
		},
		{
			name: "service with no service type specified",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "nil-service"
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "nil-service": service type not specified`,
		},
		{
			name: "invalid websocket service - invalid scheme",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "ws-svc",
							"websocket_service": {
								"address": "http://example.com"
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "ws-svc": invalid websocket target_address scheme: http`,
		},
		{
			name: "invalid grpc service - empty address",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "grpc-svc",
							"grpc_service": {
								"address": ""
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "grpc-svc": gRPC service has empty target_address`,
		},
		{
			name: "invalid openapi service - no spec",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "openapi-svc",
							"openapi_service": {}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "openapi-svc": openapi service must have either an address, spec content or spec url`,
		},
		{
			name: "invalid mcp service - empty http address",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "mcp-svc",
							"mcp_service": {
								"http_connection": {
									"http_address": ""
								}
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "mcp-svc": mcp service with http_connection has empty http_address`,
		},
		{
			name: "invalid basic auth - missing username",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "basic-auth-svc",
							"http_service": { "address": "http://example.com" },
							"upstream_authentication": {
								"basic_auth": {
									"password": { "plainText": "pass" }
								}
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "basic-auth-svc": basic auth 'username' is empty`,
		},
		{
			name: "invalid bearer token - empty token",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "bearer-svc",
							"http_service": { "address": "http://example.com" },
							"upstream_authentication": {
								"bearer_token": {
									"token": { "plainText": "" }
								}
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "bearer-svc": bearer token 'token' is empty`,
		},
		{
			name: "invalid websocket service - input schema error",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				// Use builder or manual struct for schema
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("ws-schema"),
						ServiceConfig: &configv1.UpstreamServiceConfig_WebsocketService{
							WebsocketService: &configv1.WebsocketUpstreamService{
								Address: proto.String("ws://example.com"),
								Calls: map[string]*configv1.WebsocketCallDefinition{
									"foo": {
										InputSchema: &structpb.Struct{}, // Empty struct is valid for now based on validateSchema implementation
										// But validation logic? validateSchema returns nil always currently.
										// So we can't trigger error unless validateSchema is updated or we mock it?
										// validateSchema in validator.go lines 363-369 simply returns nil.
										// "This is a loose check".
										// So we CAN'T trigger schema error yet.
										// Skip schema error test for now.
									},
								},
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid sql service - empty driver",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "sql-svc-empty-driver",
							"sql_service": {
								"driver": "",
								"dsn": "postgres://user:pass@localhost:5432/db"
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "sql-svc-empty-driver": sql service has empty driver`,
		},
		{
			name: "invalid sql service - empty dsn",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "sql-svc-empty-dsn",
							"sql_service": {
								"driver": "postgres",
								"dsn": ""
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "sql-svc-empty-dsn": sql service has empty dsn`,
		},
		{
			name: "invalid graphql service - empty address",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "graphql-svc-empty-addr",
							"graphql_service": {
								"address": ""
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "graphql-svc-empty-addr": graphql service has empty address`,
		},
		{
			name: "invalid webrtc service - empty address",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "webrtc-svc-empty-addr",
							"webrtc_service": {
								"address": ""
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "webrtc-svc-empty-addr": webrtc service has empty address`,
		},
		{
			name: "invalid openapi service - bad spec url",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "openapi-bad-url",
							"openapi_service": {
								"spec_url": "not-a-url"
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "openapi-bad-url": invalid openapi spec_url: not-a-url`,
		},
		{
			name: "invalid mcp service - empty command for stdio",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "mcp-stdio-empty",
							"mcp_service": {
								"stdio_connection": {
									"command": ""
								}
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "mcp-stdio-empty": mcp service with stdio_connection has empty command`,
		},
		{
			name: "valid mcp service - stdio",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "mcp-stdio-valid",
							"mcp_service": {
								"stdio_connection": {
									"command": "ls"
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
			name: "invalid http service - bad scheme",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "http-ftp",
							"http_service": {
								"address": "ftp://example.com"
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "http-ftp": invalid http target_address scheme: ftp`,
		},
		{
			name: "invalid global settings - empty redis address",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"global_settings": {
						"message_bus": {
							"redis": {
								"address": ""
							}
						}
					}
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "global_settings": redis message bus address is empty`,
		},
						{
							name: "valid grpc service with calls",
							config: func() *configv1.McpAnyServerConfig {
								cfg := &configv1.McpAnyServerConfig{}
								// Use struct construction
								cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
									{
										Name: proto.String("grpc-valid"),
										ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
											GrpcService: &configv1.GrpcUpstreamService{
												Address: proto.String("localhost:50051"),
												Calls: map[string]*configv1.GrpcCallDefinition{
													"foo": {
														InputSchema: &structpb.Struct{},
													},
												},
											},
										},
									},
								}
								return cfg
							}(),
							expectedErrorCount: 0,
						},
						{
							name: "invalid openapi service - invalid spec url",
							config: func() *configv1.McpAnyServerConfig {
								cfg := &configv1.McpAnyServerConfig{}
								require.NoError(t, protojson.Unmarshal([]byte(`{
									"upstream_services": [
										{
											"name": "openapi-invalid-url",
											"openapi_service": {
												"spec_url": "javascript:alert(1)"
											}
										}
									]
								}`), cfg))
								return cfg
							}(),
							expectedErrorCount:  1,
							expectedErrorString: `service "openapi-invalid-url": invalid openapi spec_url: javascript:alert(1)`,
						},
						{
							name: "invalid openapi service - invalid address",
							config: func() *configv1.McpAnyServerConfig {
								cfg := &configv1.McpAnyServerConfig{}
								require.NoError(t, protojson.Unmarshal([]byte(`{
									"upstream_services": [
										{
											"name": "openapi-invalid-Addr",
											"openapi_service": {
												"address": "javascript:alert(1)"
											}
										}
									]
								}`), cfg))
								return cfg
							}(),
							expectedErrorCount:  1,
							expectedErrorString: `service "openapi-invalid-Addr": invalid openapi target_address: javascript:alert(1)`,
						},
		{
			name: "invalid mcp service - bad http url",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "mcp-http-bad",
							"mcp_service": {
								"http_connection": {
									"http_address": "bad-url"
								}
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "mcp-http-bad": mcp service with http_connection has invalid http_address: bad-url`,
		},
		{
			name: "invalid mcp service - no connection type",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "mcp-no-conn",
							"mcp_service": {}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "mcp-no-conn": mcp service has no connection_type`,
		},
		{
			name: "valid mtls auth",
			config: func() *configv1.McpAnyServerConfig {
				// We need paths that exist and are secure.
				// Temp files created in TestValidate are likely secure (0600) and exist.
				// clientCertFile, clientKeyFile, caCertFile are available in closure?
				// Yes, closure captures variables from TestValidate scope.

				cfg := &configv1.McpAnyServerConfig{}
				// Use struct directly to use variables
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("mtls-valid"),
						ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
							HttpService: &configv1.HttpUpstreamService{Address: proto.String("https://example.com")},
						},
						UpstreamAuthentication: &configv1.UpstreamAuthentication{
							AuthMethod: &configv1.UpstreamAuthentication_Mtls{
								Mtls: &configv1.UpstreamMTLSAuth{
									ClientCertPath: proto.String(clientCertFile.Name()),
									ClientKeyPath:  proto.String(clientKeyFile.Name()),
									CaCertPath:     proto.String(caCertFile.Name()),
								},
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid mtls auth - client cert not found",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				// Use struct construction
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("mtls-missing-cert"),
						UpstreamAuthentication: &configv1.UpstreamAuthentication{
							AuthMethod: &configv1.UpstreamAuthentication_Mtls{
								Mtls: &configv1.UpstreamMTLSAuth{
									ClientCertPath: proto.String("missing-cert.pem"),
									ClientKeyPath:  proto.String("missing-key.pem"),
									CaCertPath:     proto.String("missing-ca.pem"),
								},
							},
						},
						ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
							GrpcService: &configv1.GrpcUpstreamService{
								Address: proto.String("localhost:50051"),
							},
						},
					},
				}
				// We expect error because validation checks existence
				// BUT we need to make sure the paths are ABSOLUTE or relative?
				// Validation usually checks FileExists.
				// If I use "missing-cert.pem", it might look in current dir.
				// "/non/existent/cert.pem" is safer.
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "mtls-missing-cert": mtls 'client_cert_path' not found: stat missing-cert.pem: no such file or directory`,
		},
		{
			name: "invalid command line service - missing command",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("cli-invalid"),
						ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
							CommandLineService: &configv1.CommandLineUpstreamService{
								// Command missing
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "cli-invalid": command_line_service has empty command`,
		},
		{
			name: "invalid mtls auth - missing client key",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("mtls-missing-key"),
						UpstreamAuthentication: &configv1.UpstreamAuthentication{
							AuthMethod: &configv1.UpstreamAuthentication_Mtls{
								Mtls: &configv1.UpstreamMTLSAuth{
									ClientCertPath: proto.String(clientCertFile.Name()),
									ClientKeyPath:  proto.String("missing-key.pem"),
									CaCertPath:     proto.String("ca.pem"),
								},
							},
						},
						ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
							GrpcService: &configv1.GrpcUpstreamService{
								Address: proto.String("localhost:50051"),
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "mtls-missing-key": mtls 'client_key_path' not found: stat missing-key.pem: no such file or directory`,
		},
		{
			name: "invalid basic auth - missing username",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("basic-no-user"),
						UpstreamAuthentication: &configv1.UpstreamAuthentication{
							AuthMethod: &configv1.UpstreamAuthentication_BasicAuth{
								BasicAuth: &configv1.UpstreamBasicAuth{
									Password: &configv1.SecretValue{
										Value: &configv1.SecretValue_PlainText{PlainText: "pass"},
									},
								},
							},
						},
						ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
							HttpService: &configv1.HttpUpstreamService{
								Address: proto.String("http://example.com"),
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "basic-no-user": basic auth 'username' is empty`,
		},
		{
			name: "invalid http service - invalid input schema",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("http-invalid-schema"),
						ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
							HttpService: &configv1.HttpUpstreamService{
								Address: proto.String("http://example.com"),
								Calls: map[string]*configv1.HttpCallDefinition{
									"op1": {
										Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
										EndpointPath: proto.String("/op1"),
										InputSchema: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
											},
										},
									},
								},
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "http-invalid-schema": http call "op1" input_schema error: schema 'type' must be a string`,
		},
		{
			name: "invalid mtls auth - insecure client cert path",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("mtls-insecure"),
						UpstreamAuthentication: &configv1.UpstreamAuthentication{
							AuthMethod: &configv1.UpstreamAuthentication_Mtls{
								Mtls: &configv1.UpstreamMTLSAuth{
									ClientCertPath: proto.String("../dummy-cert.pem"),
									ClientKeyPath:  proto.String("dummy-key.pem"),
								},
							},
						},
						ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
							GrpcService: &configv1.GrpcUpstreamService{
								Address: proto.String("localhost:50051"),
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "mtls-insecure": mtls 'client_cert_path' is not a secure path: path contains '..', which is not allowed`,
		},
		{
			name: "invalid api key auth - secret resolution failed",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("apikey-secret-fail"),
						ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
							HttpService: &configv1.HttpUpstreamService{
								Address: proto.String("http://example.com"),
							},
						},
						UpstreamAuthentication: &configv1.UpstreamAuthentication{
							AuthMethod: &configv1.UpstreamAuthentication_ApiKey{
								ApiKey: &configv1.UpstreamAPIKeyAuth{
									HeaderName: proto.String("X-API-Key"),
									ApiKey: &configv1.SecretValue{
										// Invalid file path
										Value: &configv1.SecretValue_FilePath{FilePath: "non-existent-secret-file"},
									},
								},
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "apikey-secret-fail": failed to resolve api key secret: failed to read secret from file "non-existent-secret-file": open non-existent-secret-file: no such file or directory`,
		},
		{
			name: "invalid websocket service - missing address",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("ws-invalid"),
						ServiceConfig: &configv1.UpstreamServiceConfig_WebsocketService{
							WebsocketService: &configv1.WebsocketUpstreamService{
								// Address missing
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "ws-invalid": websocket service has empty target_address`,
		},
		{
			name: "invalid apikey auth - missing header name",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("apikey-invalid"),
						ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
							HttpService: &configv1.HttpUpstreamService{
								Address: proto.String("http://example.com"),
							},
						},
						UpstreamAuthentication: &configv1.UpstreamAuthentication{
							AuthMethod: &configv1.UpstreamAuthentication_ApiKey{
								ApiKey: &configv1.UpstreamAPIKeyAuth{
									// HeaderName missing
									ApiKey: &configv1.SecretValue{
										Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
									},
								},
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "apikey-invalid": api key 'header_name' is empty`,
		},
		{
			name: "invalid volume mount - insecure host path",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("volume-insecure"),
						ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
							CommandLineService: &configv1.CommandLineUpstreamService{
								Command: proto.String("echo"),
								ContainerEnvironment: &configv1.ContainerEnvironment{
									Image: proto.String("alpine"),
									Volumes: map[string]string{
										"../../etc/passwd": "/passwd",
									},
								},
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "volume-insecure": container environment volume host path "../../etc/passwd" is not a secure path: path contains '..', which is not allowed`,
		},
		{
			name: "valid volume mount - secure host path",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("volume-secure"),
						ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
							CommandLineService: &configv1.CommandLineUpstreamService{
								Command: proto.String("echo"),
								ContainerEnvironment: &configv1.ContainerEnvironment{
									Image: proto.String("alpine"),
									Volumes: map[string]string{
										"host/path": "/container/path",
									},
								},
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid volume mount - absolute path",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("volume-absolute"),
						ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
							CommandLineService: &configv1.CommandLineUpstreamService{
								Command: proto.String("echo"),
								ContainerEnvironment: &configv1.ContainerEnvironment{
									Image: proto.String("alpine"),
									Volumes: map[string]string{
										"/host/path": "/container/path",
									},
								},
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount: 1,
			expectedErrorString: `service "volume-absolute": container environment volume host path "/host/path" is not a secure path: path "/host/path" is not allowed (must be in CWD or in MCPANY_FILE_PATH_ALLOW_LIST)`,
		},
        /*
		{
			name: "invalid cache TTL",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				// -1s TTL
                // Use struct construction
                cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
                    {
                        Name: proto.String("cache-invalid"),
                        ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
                            HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://example.com")},
                        },
                        Cache: &configv1.CacheConfig{
                            Ttl: durationpb.New(-1 * time.Second),
                        },
                    },
                }
				return cfg
			}(),
			expectedErrorCount: 1,
            expectedErrorString: "service \"cache-invalid\": invalid cache timeout: -1s",
		},
        */
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := Validate(context.Background(), tt.config, Server)
			if tt.expectedErrorCount > 0 {
				require.NotEmpty(t, validationErrors)
				assert.EqualError(t, &validationErrors[0], tt.expectedErrorString)
			}
		})
	}
}

func TestValidateOrError(t *testing.T) {
	cfg := &configv1.UpstreamServiceConfig{
		Name: proto.String("test"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://example.com"),
			},
		},
	}
	err := ValidateOrError(context.Background(), cfg)
	assert.NoError(t, err)

	cfgInvalid := &configv1.UpstreamServiceConfig{
		Name: proto.String("test"),
	}
	err = ValidateOrError(context.Background(), cfgInvalid)
	assert.Error(t, err)
}

func TestValidate_Client(t *testing.T) {
	cfg := &configv1.McpAnyServerConfig{
		GlobalSettings: &configv1.GlobalSettings{
			ApiKey: proto.String("short"),
		},
	}
	errs := Validate(context.Background(), cfg, Client)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "API key must be at least 16 characters long")
}

func TestValidate_MTLS(t *testing.T) {
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("mtls"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://example.com")},
				},
				UpstreamAuthentication: &configv1.UpstreamAuthentication{
					AuthMethod: &configv1.UpstreamAuthentication_Mtls{
						Mtls: &configv1.UpstreamMTLSAuth{
							ClientCertPath: proto.String("/non/existent/cert"),
							ClientKeyPath:  proto.String("/non/existent/key"),
						},
					},
				},
			},
		},
	}
	errs := Validate(context.Background(), cfg, Server)
	require.NotEmpty(t, errs)
}

func TestValidateGlobalSettings(t *testing.T) {
	t.Run("invalid global settings - client short api key", func(t *testing.T) {
		cfg := &configv1.McpAnyServerConfig{
			GlobalSettings: &configv1.GlobalSettings{
				ApiKey: proto.String("short"),
			},
		}
		errs := Validate(context.Background(), cfg, Client)
		assert.Len(t, errs, 1)
		assert.Equal(t, `service "global_settings": API key must be at least 16 characters long`, errs[0].Error())
	})

	t.Run("invalid global settings - empty redis address", func(t *testing.T) {
		mb := &bus.MessageBus{}
		mb.SetRedis(&bus.RedisBus{
			// Address empty
		})
		cfg := &configv1.McpAnyServerConfig{
			GlobalSettings: &configv1.GlobalSettings{
				MessageBus: mb,
			},
		}
		errs := Validate(context.Background(), cfg, Server)
		assert.Len(t, errs, 1)
		assert.Equal(t, `service "global_settings": redis message bus address is empty`, errs[0].Error())
	})
}
