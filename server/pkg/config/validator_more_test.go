package config

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestValidate_MoreServices(t *testing.T) {
	// Mock execLookPath
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		if file == "ls" {
			return "/bin/ls", nil
		}
		return "", os.ErrNotExist
	}

	// Create an insecure file (world writable)
	insecureFile, err := os.CreateTemp("", "insecure_key")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(insecureFile.Name()) }()
	if err := insecureFile.Chmod(0777); err != nil {
		t.Fatal(err)
	}
	insecurePath := insecureFile.Name()

	// Allow temp directory for tests
	validation.SetAllowedPaths([]string{os.TempDir()})
	defer validation.SetAllowedPaths(nil)

	tests := []struct {
		name                string
		config              *configv1.McpAnyServerConfig
		expectedErrorCount  int
		expectedErrorString string
	}{
		{
			name: "MTLS Auth Client Key Path Not Secure",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mtls-service-insecure-key"),
						HttpService: configv1.HttpUpstreamService_builder{
							Address: proto.String("http://example.com"),
						}.Build(),
						UpstreamAuth: configv1.Authentication_builder{
							Mtls: configv1.MTLSAuth_builder{
								ClientCertPath: proto.String(insecurePath),
								ClientKeyPath:  proto.String(insecurePath),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  0,
			expectedErrorString: "",
		},
		{
			name: "MTLS Auth Unauthorized Path",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mtls-unauthorized"),
						HttpService: configv1.HttpUpstreamService_builder{
							Address: proto.String("http://example.com"),
						}.Build(),
						UpstreamAuth: configv1.Authentication_builder{
							Mtls: configv1.MTLSAuth_builder{
								ClientCertPath: proto.String("/etc/passwd"),
								ClientKeyPath:  proto.String("/etc/passwd"),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "mtls 'client_cert_path' is not a secure path",
		},
		{
			name: "MTLS Auth File Not Found",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mtls-service-missing-file"),
						HttpService: configv1.HttpUpstreamService_builder{
							Address: proto.String("http://example.com"),
						}.Build(),
						UpstreamAuth: configv1.Authentication_builder{
							Mtls: configv1.MTLSAuth_builder{
								ClientCertPath: proto.String("non-existent-cert.pem"),
								ClientKeyPath:  proto.String("non-existent-key.pem"),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "mtls 'client_cert_path' not found",
		},
		{
			name: "Basic Auth Unset Env",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("basic-auth-unset-env"),
						HttpService: configv1.HttpUpstreamService_builder{
							Address: proto.String("http://example.com"),
						}.Build(),
						UpstreamAuth: configv1.Authentication_builder{
							BasicAuth: configv1.BasicAuth_builder{
								Username: proto.String("user"),
								Password: configv1.SecretValue_builder{
									EnvironmentVariable: proto.String("UNSET_ENV_VAR_FOR_TEST"),
								}.Build(),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "basic auth password validation failed",
		},
		{
			name: "Bearer Token Unset Env",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("bearer-auth-unset-env"),
						HttpService: configv1.HttpUpstreamService_builder{
							Address: proto.String("http://example.com"),
						}.Build(),
						UpstreamAuth: configv1.Authentication_builder{
							BearerToken: configv1.BearerTokenAuth_builder{
								Token: configv1.SecretValue_builder{
									EnvironmentVariable: proto.String("UNSET_ENV_VAR_FOR_TEST"),
								}.Build(),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "bearer token validation failed",
		},
		{
			name: "invalid command line service - missing command",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name:               proto.String("cmd-invalid"),
						CommandLineService: configv1.CommandLineUpstreamService_builder{}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "cmd-invalid": command_line_service has empty command`,
		},
		{
			name: "invalid grpc service - missing address",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name:        proto.String("grpc-invalid"),
						GrpcService: configv1.GrpcUpstreamService_builder{}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "grpc-invalid": gRPC service has empty address`,
		},
		{
			name: "invalid mcp service - missing connection",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name:       proto.String("mcp-invalid"),
						McpService: configv1.McpUpstreamService_builder{}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "mcp-invalid": mcp service has no connection_type`,
		},
		{
			name: "invalid mcp service - invalid stdio connection",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mcp-stdio-invalid"),
						McpService: configv1.McpUpstreamService_builder{
							StdioConnection: configv1.McpStdioConnection_builder{}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "mcp-stdio-invalid": mcp service with stdio_connection has empty command`,
		},
		{
			name: "invalid mcp service - invalid http connection",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mcp-http-invalid"),
						McpService: configv1.McpUpstreamService_builder{
							HttpConnection: configv1.McpStreamableHttpConnection_builder{}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: `service "mcp-http-invalid": mcp service with http_connection has empty http_address`,
		},
		{
			name: "invalid http service - empty address",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("http-empty"),
						HttpService: configv1.HttpUpstreamService_builder{}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "http service has empty address",
		},
		{
			name: "invalid http service - invalid url",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("http-invalid-url"),
						HttpService: configv1.HttpUpstreamService_builder{
							Address: proto.String("::invalid-url"),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "invalid http address",
		},
		{
			name: "invalid http service - invalid scheme",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("http-invalid-scheme"),
						HttpService: configv1.HttpUpstreamService_builder{
							Address: proto.String("ftp://example.com"),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "invalid http address scheme",
		},
		{
			name: "invalid websocket service - empty address",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name:             proto.String("ws-empty"),
						WebsocketService: configv1.WebsocketUpstreamService_builder{}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "websocket service has empty address",
		},
		{
			name: "invalid websocket service - invalid scheme",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("ws-invalid-scheme"),
						WebsocketService: configv1.WebsocketUpstreamService_builder{
							Address: proto.String("http://example.com"),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "invalid websocket address scheme",
		},
		{
			name: "invalid openapi service - all empty",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name:           proto.String("openapi-empty"),
						OpenapiService: configv1.OpenapiUpstreamService_builder{}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "openapi service must have either an address, spec content or spec url",
		},
		{
			name: "invalid openapi service - invalid address",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("openapi-invalid-addr"),
						OpenapiService: configv1.OpenapiUpstreamService_builder{
							Address: proto.String("::invalid"),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "invalid openapi address",
		},
		{
			name: "invalid openapi service - invalid spec url",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("openapi-invalid-spec-url"),
						OpenapiService: configv1.OpenapiUpstreamService_builder{
							SpecUrl: proto.String("::invalid"),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "invalid openapi spec_url",
		},
		{
			name: "valid command line service",
			config: func() *configv1.McpAnyServerConfig {
				cmdSvc := configv1.CommandLineUpstreamService_builder{
					Command: proto.String("ls"),
				}.Build()

				svc := configv1.UpstreamServiceConfig_builder{
					Name:               proto.String("cmd-valid"),
					CommandLineService: cmdSvc,
				}.Build()

				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
				}.Build()
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid apikey auth - empty key value",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("apikey-empty-val"),
						HttpService: configv1.HttpUpstreamService_builder{
							Address: proto.String("http://example.com"),
						}.Build(),
						UpstreamAuth: configv1.Authentication_builder{
							ApiKey: configv1.APIKeyAuth_builder{
								ParamName: proto.String("X-API-Key"),
								Value: configv1.SecretValue_builder{
									PlainText: proto.String(""),
								}.Build(),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "resolved api key value is empty",
		},
		{
			name: "invalid basic auth - empty password",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("basic-empty-pass"),
						HttpService: configv1.HttpUpstreamService_builder{
							Address: proto.String("http://example.com"),
						}.Build(),
						UpstreamAuth: configv1.Authentication_builder{
							BasicAuth: configv1.BasicAuth_builder{
								Username: proto.String("user"),
								Password: configv1.SecretValue_builder{
									PlainText: proto.String(""),
								}.Build(),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "basic auth 'password' is empty",
		},
		{
			name: "valid mtls auth - no ca",
			config: func() *configv1.McpAnyServerConfig {
				mtls := configv1.MTLSAuth_builder{
					ClientCertPath: proto.String(insecurePath),
					ClientKeyPath:  proto.String(insecurePath),
				}.Build()

				auth := configv1.Authentication_builder{
					Mtls: mtls,
				}.Build()

				svc := configv1.UpstreamServiceConfig_builder{
					Name: proto.String("mtls-valid-no-ca"),
					HttpService: configv1.HttpUpstreamService_builder{
						Address: proto.String("https://example.com"),
					}.Build(),
					UpstreamAuth: auth,
				}.Build()

				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
				}.Build()
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "empty upstream auth config",
			config: func() *configv1.McpAnyServerConfig {
				auth := configv1.Authentication_builder{}.Build()

				svc := configv1.UpstreamServiceConfig_builder{
					Name: proto.String("empty-auth"),
					HttpService: configv1.HttpUpstreamService_builder{
						Address: proto.String("https://example.com"),
					}.Build(),
					UpstreamAuth: auth,
				}.Build()

				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
				}.Build()
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid mcp service - call schema error",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mcp-schema-error"),
						McpService: configv1.McpUpstreamService_builder{
							StdioConnection: configv1.McpStdioConnection_builder{
								Command: proto.String("ls"),
							}.Build(),
							Calls: map[string]*configv1.MCPCallDefinition{
								"bad-call": configv1.MCPCallDefinition_builder{
									InputSchema: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
										},
									},
								}.Build(),
							},
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "input_schema error",
		},
		{
			name: "mtls auth insecure ca cert",
			config: func() *configv1.McpAnyServerConfig {
				svc := configv1.UpstreamServiceConfig_builder{
					Name: proto.String("mtls-insecure-ca"),
					HttpService: configv1.HttpUpstreamService_builder{
						Address: proto.String("https://example.com"),
					}.Build(),
					UpstreamAuth: configv1.Authentication_builder{
						Mtls: configv1.MTLSAuth_builder{
							ClientCertPath: proto.String(insecurePath),
							ClientKeyPath:  proto.String(insecurePath),
							CaCertPath:     proto.String(insecurePath),
						}.Build(),
					}.Build(),
				}.Build()

				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
				}.Build()
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "mtls auth missing ca cert",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mtls-missing-ca"),
						HttpService: configv1.HttpUpstreamService_builder{
							Address: proto.String("https://example.com"),
						}.Build(),
						UpstreamAuth: configv1.Authentication_builder{
							Mtls: configv1.MTLSAuth_builder{
								ClientCertPath: proto.String(insecurePath),
								ClientKeyPath:  proto.String(insecurePath),
								CaCertPath:     proto.String("non-existent-ca.pem"),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "mtls 'ca_cert_path' not found",
		},
		{
			name: "invalid grpc service - schema error",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("grpc-schema-error"),
						GrpcService: configv1.GrpcUpstreamService_builder{
							Address: proto.String("127.0.0.1:50051"),
							Calls: map[string]*configv1.GrpcCallDefinition{
								"bad-call": configv1.GrpcCallDefinition_builder{
									InputSchema: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
										},
									},
								}.Build(),
							},
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "input_schema error",
		},
		{
			name: "invalid websocket service - call input schema error",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("ws-input-error"),
						WebsocketService: configv1.WebsocketUpstreamService_builder{
							Address: proto.String("ws://example.com"),
							Calls: map[string]*configv1.WebsocketCallDefinition{
								"bad-input": configv1.WebsocketCallDefinition_builder{
									InputSchema: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
										},
									},
								}.Build(),
							},
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "input_schema error",
		},
		{
			name: "invalid mcp service - call output schema error",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mcp-output-error"),
						McpService: configv1.McpUpstreamService_builder{
							StdioConnection: configv1.McpStdioConnection_builder{
								Command: proto.String("ls"),
							}.Build(),
							Calls: map[string]*configv1.MCPCallDefinition{
								"bad-output": configv1.MCPCallDefinition_builder{
									OutputSchema: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
										},
									},
								}.Build(),
							},
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "output_schema error",
		},
		{
			name: "invalid websocket service - call output schema error",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("ws-output-error"),
						WebsocketService: configv1.WebsocketUpstreamService_builder{
							Address: proto.String("ws://example.com"),
							Calls: map[string]*configv1.WebsocketCallDefinition{
								"bad-output": configv1.WebsocketCallDefinition_builder{
									OutputSchema: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
										},
									},
								}.Build(),
							},
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "output_schema error",
		},
		{
			name: "valid openapi service - spec content",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("openapi-content"),
						OpenapiService: configv1.OpenapiUpstreamService_builder{
							SpecContent: proto.String(`{"openapi": "3.0.0"}`),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid grpc service - output schema error",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("grpc-output-error"),
						GrpcService: configv1.GrpcUpstreamService_builder{
							Address: proto.String("127.0.0.1:50051"),
							Calls: map[string]*configv1.GrpcCallDefinition{
								"bad-output": configv1.GrpcCallDefinition_builder{
									OutputSchema: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
										},
									},
								}.Build(),
							},
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "output_schema error",
		},
		{
			name: "invalid websocket service - call input schema error",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("ws-input-error"),
						WebsocketService: configv1.WebsocketUpstreamService_builder{
							Address: proto.String("ws://example.com"),
							Calls: map[string]*configv1.WebsocketCallDefinition{
								"bad-input": configv1.WebsocketCallDefinition_builder{
									InputSchema: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
										},
									},
								}.Build(),
							},
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "input_schema error",
		},
		{
			name: "invalid mcp service - call output schema error",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mcp-output-error"),
						McpService: configv1.McpUpstreamService_builder{
							StdioConnection: configv1.McpStdioConnection_builder{
								Command: proto.String("ls"),
							}.Build(),
							Calls: map[string]*configv1.MCPCallDefinition{
								"bad-output": configv1.MCPCallDefinition_builder{
									OutputSchema: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
										},
									},
								}.Build(),
							},
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "output_schema error",
		},
		{
			name: "invalid websocket service - call output schema error",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("ws-output-error"),
						WebsocketService: configv1.WebsocketUpstreamService_builder{
							Address: proto.String("ws://example.com"),
							Calls: map[string]*configv1.WebsocketCallDefinition{
								"bad-output": configv1.WebsocketCallDefinition_builder{
									OutputSchema: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
										},
									},
								}.Build(),
							},
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "output_schema error",
		},
		{
			name: "valid openapi service - spec content",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("openapi-content"),
						OpenapiService: configv1.OpenapiUpstreamService_builder{
							SpecContent: proto.String(`{"openapi": "3.0.0"}`),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid grpc service - output schema error",
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("grpc-output-error"),
						GrpcService: configv1.GrpcUpstreamService_builder{
							Address: proto.String("127.0.0.1:50051"),
							Calls: map[string]*configv1.GrpcCallDefinition{
								"bad-output": configv1.GrpcCallDefinition_builder{
									OutputSchema: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
										},
									},
								}.Build(),
							},
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "output_schema error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := Validate(context.Background(), tc.config, Server)
			assert.Equal(t, tc.expectedErrorCount, len(errs), "error count mismatch for %s", tc.name)
			if tc.expectedErrorString != "" && len(errs) > 0 {
				assert.Contains(t, errs[0].Error(), tc.expectedErrorString)
			}
		})
	}
}

func TestValidate_MtlsInsecure(t *testing.T) {
	// Create a real file for insecure path
	f, err := os.CreateTemp("", "insecure.pem")
	require.NoError(t, err)
	defer func() { _ = os.Remove(f.Name()) }()
	insecurePath := f.Name()

	// Create unreadable directory to force os.Stat error
	dirPerm, err := os.MkdirTemp("", "unreadable_dir")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(dirPerm) }()

	// Make it unreadable
	if err := os.Chmod(dirPerm, 0000); err != nil {
		t.Fatalf("Failed to chmod: %v", err)
	}
	//nolint:gosec // Directory permissions need to be 0755
	defer func() { _ = os.Chmod(dirPerm, 0755) }()

	// Create a real file for secure path (needs to exist for FileExists check)
	fSecure, err := os.CreateTemp("", "secure.pem")
	require.NoError(t, err)
	defer func() { _ = os.Remove(fSecure.Name()) }()
	_ = fSecure.Close()
	securePath := fSecure.Name()

	// Mock IsAllowedPath
	originalIsAllowedPath := validation.IsAllowedPath
	validation.IsAllowedPath = func(path string) error {
		if path == "insecure.pem" || path == insecurePath {
			return fmt.Errorf("mock insecure path")
		}
		// All other paths are secure
		return nil
	}
	defer func() { validation.IsAllowedPath = originalIsAllowedPath }()

	// Mock osStat
	originalOsStat := osStat
	osStat = func(name string) (os.FileInfo, error) {
		if name == "mock-error-path" {
			return nil, fmt.Errorf("mock error")
		}
		return originalOsStat(name)
	}
	defer func() { osStat = originalOsStat }()

	tests := []struct {
		name                string
		config              *configv1.McpAnyServerConfig
		expectedErrorCount  int
		expectedErrorString string
	}{
		{
			name: "mtls auth - insecure ca cert mocked",
			config: func() *configv1.McpAnyServerConfig {
				mtls := configv1.MTLSAuth_builder{
					ClientCertPath: proto.String(securePath),
					ClientKeyPath:  proto.String(securePath),
					CaCertPath:     proto.String("insecure.pem"),
				}.Build()

				auth := configv1.Authentication_builder{
					Mtls: mtls,
				}.Build()

				svc := configv1.UpstreamServiceConfig_builder{
					Name: proto.String("mtls-mock-insecure-ca"),
					HttpService: configv1.HttpUpstreamService_builder{
						Address: proto.String("https://example.com"),
					}.Build(),
					UpstreamAuth: auth,
				}.Build()

				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
				}.Build()
			}(),
			expectedErrorCount:  1,
			expectedErrorString: "mtls 'ca_cert_path' is not a secure path",
		},
		{
			name: "mtls auth - insecure client cert mocked",
			config: func() *configv1.McpAnyServerConfig {
				mtls := configv1.MTLSAuth_builder{
					ClientCertPath: proto.String("insecure.pem"),
					ClientKeyPath:  proto.String(securePath),
				}.Build()

				auth := configv1.Authentication_builder{
					Mtls: mtls,
				}.Build()

				svc := configv1.UpstreamServiceConfig_builder{
					Name: proto.String("mtls-mock-insecure-client"),
					HttpService: configv1.HttpUpstreamService_builder{
						Address: proto.String("https://example.com"),
					}.Build(),
					UpstreamAuth: auth,
				}.Build()

				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
				}.Build()
			}(),
			expectedErrorCount:  1,
			expectedErrorString: "mtls 'client_cert_path' is not a secure path",
		},
		{
			name: "mtls auth - insecure client key mocked",
			// Test case: Client Key is world readable (insecure)
			config: configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("mtls-insecure-key-file"),
						HttpService: configv1.HttpUpstreamService_builder{
							Address: proto.String("https://example.com"),
						}.Build(),
						UpstreamAuth: configv1.Authentication_builder{
							Mtls: configv1.MTLSAuth_builder{
								ClientCertPath: proto.String(securePath),
								ClientKeyPath:  proto.String(insecurePath),
							}.Build(),
						}.Build(),
					}.Build(),
				},
			}.Build(),
			expectedErrorCount:  1,
			expectedErrorString: "mtls 'client_key_path' is not a secure path",
		},
		{
			name: "mtls auth - valid path but missing file",
			config: func() *configv1.McpAnyServerConfig {
				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{
						configv1.UpstreamServiceConfig_builder{
							Name: proto.String("mtls-missing-file"),
							HttpService: configv1.HttpUpstreamService_builder{
								Address: proto.String("https://example.com"),
							}.Build(),
							UpstreamAuth: configv1.Authentication_builder{
								Mtls: configv1.MTLSAuth_builder{
									ClientCertPath: proto.String(securePath),
									ClientKeyPath:  proto.String(securePath + ".missing"),
								}.Build(),
							}.Build(),
						}.Build(),
					},
				}.Build()
			}(),
			expectedErrorCount:  1,
			expectedErrorString: "mtls 'client_key_path' not found",
		},
		{
			name: "mtls auth - permission denied",
			config: func() *configv1.McpAnyServerConfig {
				return configv1.McpAnyServerConfig_builder{
					UpstreamServices: []*configv1.UpstreamServiceConfig{
						configv1.UpstreamServiceConfig_builder{
							Name: proto.String("mtls-permission-denied"),
							HttpService: configv1.HttpUpstreamService_builder{
								Address: proto.String("https://example.com"),
							}.Build(),
							UpstreamAuth: configv1.Authentication_builder{
								Mtls: configv1.MTLSAuth_builder{
									ClientCertPath: proto.String(securePath),
									ClientKeyPath:  proto.String("mock-error-path"),
								}.Build(),
							}.Build(),
						}.Build(),
					},
				}.Build()
			}(),
			expectedErrorCount:  1,
			expectedErrorString: "mtls 'client_key_path' error", // Matches "error: mock error"
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// We need to ensure "insecure.pem" exists if we want to pass IsSecurePath FIRST?
			// IsSecurePath is checked BEFORE FileExists.
			// So if IsSecurePath fails, we get error immediately.
			// "insecure.pem" created as temp file in TestValidate_MtlsInsecure setup.

			errs := Validate(context.Background(), tc.config, Server)
			if tc.expectedErrorCount != len(errs) {
				t.Fatalf("error count mismatch: expected %d, got %d. Errors: %v", tc.expectedErrorCount, len(errs), errs)
			}
			if tc.expectedErrorString != "" && len(errs) > 0 {
				assert.Contains(t, errs[0].Error(), tc.expectedErrorString)
			}
		})
	}
}

func TestValidate_RedisAddressNil(t *testing.T) {
	msgBus := &bus.MessageBus{}
	msgBus.SetRedis(&bus.RedisBus{})

	config := configv1.McpAnyServerConfig_builder{
		GlobalSettings: configv1.GlobalSettings_builder{
			MessageBus: msgBus,
		}.Build(),
	}.Build()
	// Address is empty/nil -> GetAddress() == "" -> Error
	errs := Validate(context.Background(), config, Server)
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "redis message bus address is empty")
}

func TestValidate_MemoryBus(t *testing.T) {
	msgBus := &bus.MessageBus{}
	msgBus.SetInMemory(&bus.InMemoryBus{})

	config := configv1.McpAnyServerConfig_builder{
		GlobalSettings: configv1.GlobalSettings_builder{
			MessageBus: msgBus,
		}.Build(),
	}.Build()
	errs := Validate(context.Background(), config, Server)
	assert.Empty(t, errs)
}
