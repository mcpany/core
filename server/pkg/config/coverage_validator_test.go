// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestValidateSchema_Coverage(t *testing.T) {
	tests := []struct {
		name      string
		schema    *structpb.Struct
		expectErr string
	}{
		{
			name: "type_not_string",
			schema: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
				},
			},
			expectErr: "schema 'type' must be a string",
		},
        {
            name: "nil_schema",
            schema: nil,
            expectErr: "",
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchema(tt.schema)
			if tt.expectErr != "" {
				assert.ErrorContains(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCommandExists_Coverage(t *testing.T) {
	// Create a temporary directory and file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-exec")
	os.WriteFile(tmpFile, []byte("#!/bin/sh\nexit 0"), 0755)

	tmpDirExec := filepath.Join(tmpDir, "dir-exec")
	os.Mkdir(tmpDirExec, 0755)

	tests := []struct {
		name       string
		command    string
		workingDir string
		expectErr  string
	}{
		{
			name:       "relative_path_exists_in_cwd",
			command:    "./test-exec",
			workingDir: tmpDir,
			expectErr:  "",
		},
		{
			name:       "relative_path_is_dir",
			command:    "./dir-exec",
			workingDir: tmpDir,
			expectErr:  "is a directory, not an executable",
		},
        {
            name: "absolute_path_exists",
            command: tmpFile,
            workingDir: "",
            expectErr: "",
        },
        {
            name: "absolute_path_not_exists",
            command: filepath.Join(tmpDir, "missing"),
            workingDir: "",
            expectErr: "executable not found",
        },
        {
            name: "absolute_path_is_dir",
            command: tmpDirExec,
            workingDir: "",
            expectErr: "is a directory, not an executable",
        },
        {
            name: "command_not_in_path",
            command: "nonexistentcommand123",
            workingDir: "",
            expectErr: "command \"nonexistentcommand123\" not found in PATH",
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommandExists(tt.command, tt.workingDir)
			if tt.expectErr != "" {
				assert.ErrorContains(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDirectoryExists_Coverage(t *testing.T) {
    tmpDir := t.TempDir()
    tmpFile := filepath.Join(tmpDir, "test-file")
    os.WriteFile(tmpFile, []byte("content"), 0644)

    tests := []struct {
        name string
        path string
        expectErr string
    }{
        {
            name: "path_is_not_dir",
            path: tmpFile,
            expectErr: "is not a directory",
        },
        {
            name: "path_not_exist",
            path: filepath.Join(tmpDir, "non-existent"),
            expectErr: "does not exist",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateDirectoryExists(tt.path)
            if tt.expectErr != "" {
                assert.ErrorContains(t, err, tt.expectErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestValidateFileExists_Coverage(t *testing.T) {
    tmpDir := t.TempDir()
    tmpFile := filepath.Join(tmpDir, "test-file")
    os.WriteFile(tmpFile, []byte("content"), 0644)

    tests := []struct {
        name string
        path string
        workingDir string
        expectErr string
    }{
        {
            name: "valid_file",
            path: tmpFile,
            workingDir: "",
            expectErr: "",
        },
        {
            name: "is_directory",
            path: tmpDir,
            workingDir: "",
            expectErr: "is a directory, expected a file",
        },
        {
            name: "does_not_exist",
            path: "missing-file",
            workingDir: tmpDir,
            expectErr: "file not found",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateFileExists(tt.path, tt.workingDir)
            if tt.expectErr != "" {
                assert.ErrorContains(t, err, tt.expectErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestValidateGCSettings_Coverage(t *testing.T) {
    tests := []struct {
        name string
        gc *configv1.GCSettings
        expectErr string
    }{
        {
            name: "invalid_interval",
            gc: &configv1.GCSettings{
                Interval: proto.String("invalid"),
            },
            expectErr: "invalid interval",
        },
        {
            name: "invalid_ttl",
            gc: &configv1.GCSettings{
                Ttl: proto.String("invalid"),
            },
            expectErr: "invalid ttl",
        },
        {
            name: "empty_gc_path",
            gc: &configv1.GCSettings{
                Enabled: proto.Bool(true),
                Paths:   []string{""},
            },
            expectErr: "empty gc path",
        },
        {
            name: "relative_gc_path",
            gc: &configv1.GCSettings{
                Enabled: proto.Bool(true),
                Paths:   []string{"relative/path"},
            },
            expectErr: "gc path \"relative/path\" must be absolute",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateGCSettings(tt.gc)
            if tt.expectErr != "" {
                assert.ErrorContains(t, err, tt.expectErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestValidateMcpService_Coverage(t *testing.T) {
    tests := []struct {
        name string
        service *configv1.McpUpstreamService
        expectErr string
    }{
        {
            name: "http_empty_address",
            service: &configv1.McpUpstreamService{
                ConnectionType: &configv1.McpUpstreamService_HttpConnection{
                    HttpConnection: &configv1.McpStreamableHttpConnection{
                        HttpAddress: proto.String(""),
                    },
                },
            },
            expectErr: "mcp service with http_connection has empty http_address",
        },
        {
            name: "http_invalid_address",
            service: &configv1.McpUpstreamService{
                ConnectionType: &configv1.McpUpstreamService_HttpConnection{
                    HttpConnection: &configv1.McpStreamableHttpConnection{
                        HttpAddress: proto.String("not-a-url"),
                    },
                },
            },
            expectErr: "invalid http_address",
        },
        {
            name: "stdio_empty_command",
            service: &configv1.McpUpstreamService{
                ConnectionType: &configv1.McpUpstreamService_StdioConnection{
                    StdioConnection: &configv1.McpStdioConnection{
                        Command: proto.String(""),
                    },
                },
            },
            expectErr: "mcp service with stdio_connection has empty command",
        },
        {
            name: "bundle_empty_path",
            service: &configv1.McpUpstreamService{
                ConnectionType: &configv1.McpUpstreamService_BundleConnection{
                    BundleConnection: &configv1.McpBundleConnection{
                        BundlePath: proto.String(""),
                    },
                },
            },
            expectErr: "mcp service with bundle_connection has empty bundle_path",
        },
        {
            name: "bundle_insecure_path",
            service: &configv1.McpUpstreamService{
                ConnectionType: &configv1.McpUpstreamService_BundleConnection{
                    BundleConnection: &configv1.McpBundleConnection{
                        BundlePath: proto.String("/etc/passwd"),
                    },
                },
            },
            expectErr: "mcp service with bundle_connection has insecure bundle_path",
        },
        {
            name: "bundle_invalid_env",
            service: &configv1.McpUpstreamService{
                ConnectionType: &configv1.McpUpstreamService_BundleConnection{
                    BundleConnection: &configv1.McpBundleConnection{
                        BundlePath: proto.String("./bundle"),
                        Env: map[string]*configv1.SecretValue{
                            "KEY": {Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MISSING"}},
                        },
                    },
                },
            },
            expectErr: "mcp service with bundle_connection has invalid secret environment variable",
        },
        {
            name: "unknown_connection_type",
            service: &configv1.McpUpstreamService{},
            expectErr: "mcp service has no connection_type",
        },
        {
            name: "input_schema_error",
            service: &configv1.McpUpstreamService{
                ConnectionType: &configv1.McpUpstreamService_HttpConnection{
                    HttpConnection: &configv1.McpStreamableHttpConnection{
                        HttpAddress: proto.String("http://example.com"),
                    },
                },
                Calls: map[string]*configv1.MCPCallDefinition{
                    "test": {
                        InputSchema: &structpb.Struct{
                            Fields: map[string]*structpb.Value{
                                "type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
                            },
                        },
                    },
                },
            },
            expectErr: "input_schema error",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateMcpService(context.Background(), tt.service)
            if tt.expectErr != "" {
                assert.ErrorContains(t, err, tt.expectErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestValidateHTTPService_Coverage(t *testing.T) {
    tests := []struct {
        name string
        service *configv1.HttpUpstreamService
        expectErr string
    }{
        {
            name: "empty_address",
            service: &configv1.HttpUpstreamService{
                Address: proto.String(""),
            },
            expectErr: "http service has empty address",
        },
        {
            name: "invalid_address",
            service: &configv1.HttpUpstreamService{
                Address: proto.String("not-a-url"),
            },
            expectErr: "invalid http address", // Depending on url.Parse implementation, might pass validation.IsValidURL but fail scheme check? validation.IsValidURL checks if it parses.
        },
        {
            name: "invalid_scheme",
            service: &configv1.HttpUpstreamService{
                Address: proto.String("ftp://example.com"),
            },
            expectErr: "invalid http address scheme",
        },
        {
            name: "input_schema_error",
            service: &configv1.HttpUpstreamService{
                Address: proto.String("http://example.com"),
                Calls: map[string]*configv1.HttpCallDefinition{
                    "test": {
                        InputSchema: &structpb.Struct{
                            Fields: map[string]*structpb.Value{
                                "type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
                            },
                        },
                    },
                },
            },
            expectErr: "input_schema error",
        },
        {
            name: "output_schema_error",
            service: &configv1.HttpUpstreamService{
                Address: proto.String("http://example.com"),
                Calls: map[string]*configv1.HttpCallDefinition{
                    "test": {
                        OutputSchema: &structpb.Struct{
                            Fields: map[string]*structpb.Value{
                                "type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
                            },
                        },
                    },
                },
            },
            expectErr: "output_schema error",
        },
        {
            name: "valid_service",
            service: &configv1.HttpUpstreamService{
                Address: proto.String("http://example.com"),
                Calls: map[string]*configv1.HttpCallDefinition{
                    "test": {
                        InputSchema: &structpb.Struct{
                            Fields: map[string]*structpb.Value{
                                "type": {Kind: &structpb.Value_StringValue{StringValue: "object"}},
                            },
                        },
                    },
                },
            },
            expectErr: "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateHTTPService(tt.service)
             if tt.expectErr != "" {
                assert.ErrorContains(t, err, tt.expectErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestValidateWebSocketService_Coverage(t *testing.T) {
     tests := []struct {
        name string
        service *configv1.WebsocketUpstreamService
        expectErr string
    }{
        {
            name: "empty_address",
            service: &configv1.WebsocketUpstreamService{
                Address: proto.String(""),
            },
            expectErr: "websocket service has empty address",
        },
        {
             name: "invalid_scheme",
            service: &configv1.WebsocketUpstreamService{
                Address: proto.String("http://example.com"),
            },
            expectErr: "invalid websocket address scheme",
        },
        {
            name: "input_schema_error",
            service: &configv1.WebsocketUpstreamService{
                Address: proto.String("ws://example.com"),
                Calls: map[string]*configv1.WebsocketCallDefinition{
                    "test": {
                        InputSchema: &structpb.Struct{
                            Fields: map[string]*structpb.Value{
                                "type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
                            },
                        },
                    },
                },
            },
            expectErr: "input_schema error",
        },
          {
            name: "output_schema_error",
            service: &configv1.WebsocketUpstreamService{
                Address: proto.String("ws://example.com"),
                Calls: map[string]*configv1.WebsocketCallDefinition{
                    "test": {
                        OutputSchema: &structpb.Struct{
                            Fields: map[string]*structpb.Value{
                                "type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
                            },
                        },
                    },
                },
            },
            expectErr: "output_schema error",
        },
         {
            name: "valid_service",
            service: &configv1.WebsocketUpstreamService{
                Address: proto.String("ws://example.com"),
                Calls: map[string]*configv1.WebsocketCallDefinition{
                    "test": {
                        InputSchema: &structpb.Struct{
                            Fields: map[string]*structpb.Value{
                                "type": {Kind: &structpb.Value_StringValue{StringValue: "object"}},
                            },
                        },
                    },
                },
            },
            expectErr: "",
        },
    }

     for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateWebSocketService(tt.service)
             if tt.expectErr != "" {
                assert.ErrorContains(t, err, tt.expectErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestValidateContainerEnvironment_Coverage(t *testing.T) {
    tests := []struct {
        name string
        env *configv1.ContainerEnvironment
        expectErr string
    }{
        {
            name: "valid_no_image",
            env: nil,
            expectErr: "",
        },
        {
            name: "empty_host_path",
            env: &configv1.ContainerEnvironment{
                Image: proto.String("alpine"),
                Volumes: map[string]string{
                    "": "/container/path",
                },
            },
            expectErr: "container environment volume host path is empty",
        },
        {
            name: "empty_container_path",
            env: &configv1.ContainerEnvironment{
                Image: proto.String("alpine"),
                Volumes: map[string]string{
                    "/host/path": "",
                },
            },
            expectErr: "container environment volume container path is empty",
        },
        {
            name: "insecure_host_path",
            env: &configv1.ContainerEnvironment{
                Image: proto.String("alpine"),
                Volumes: map[string]string{
                    "/etc/passwd": "/container/passwd",
                },
            },
            expectErr: "not a secure path",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
			err := validateContainerEnvironment(context.Background(), tt.env)
             if tt.expectErr != "" {
                assert.ErrorContains(t, err, tt.expectErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestValidateAPIKeyAuth_Coverage(t *testing.T) {
    ctx := context.Background()
    tests := []struct {
        name string
        apiKey *configv1.APIKeyAuth
        ctxType AuthValidationContext
        expectErr string
    }{
        {
            name: "empty_param_name",
            apiKey: &configv1.APIKeyAuth{
                ParamName: proto.String(""),
            },
            ctxType: AuthValidationContextIncoming,
            expectErr: "param_name",
        },
        {
            name: "outgoing_missing_value",
            apiKey: &configv1.APIKeyAuth{
                ParamName: proto.String("api_key"),
                Value: nil,
            },
            ctxType: AuthValidationContextOutgoing,
            expectErr: "api key 'value' is missing",
        },
        {
            name: "incoming_missing_config",
            apiKey: &configv1.APIKeyAuth{
                ParamName: proto.String("api_key"),
                Value: nil,
                VerificationValue: proto.String(""),
            },
            ctxType: AuthValidationContextIncoming,
            expectErr: "api key configuration is empty",
        },
        {
             name: "invalid_secret_value",
             apiKey: &configv1.APIKeyAuth{
                 ParamName: proto.String("api_key"),
                 Value: &configv1.SecretValue{Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MISSING_VAR"}},
             },
             ctxType: AuthValidationContextOutgoing,
             expectErr: "api key secret validation failed",
        },
        {
            name: "valid_incoming_api_key",
            apiKey: &configv1.APIKeyAuth{
                ParamName: proto.String("api_key"),
                VerificationValue: proto.String("secret"),
            },
            ctxType: AuthValidationContextIncoming,
            expectErr: "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateAPIKeyAuth(ctx, tt.apiKey, tt.ctxType)
            if tt.expectErr != "" {
                assert.ErrorContains(t, err, tt.expectErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestValidateBearerTokenAuth_Coverage(t *testing.T) {
    ctx := context.Background()
    tests := []struct {
        name string
        token *configv1.BearerTokenAuth
        expectErr string
    }{
         {
             name: "invalid_secret",
             token: &configv1.BearerTokenAuth{
                 Token: &configv1.SecretValue{Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MISSING_VAR"}},
             },
             expectErr: "bearer token validation failed",
         },
         {
             name: "empty_secret_resolved",
             token: &configv1.BearerTokenAuth{
                 Token: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: ""}},
             },
             expectErr: "bearer token 'token' is empty",
         },
         {
             name: "valid_bearer_token",
             token: &configv1.BearerTokenAuth{
                 Token: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "token"}},
             },
             expectErr: "",
         },
    }
     for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateBearerTokenAuth(ctx, tt.token)
            if tt.expectErr != "" {
                assert.ErrorContains(t, err, tt.expectErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestValidateBasicAuth_Coverage(t *testing.T) {
    ctx := context.Background()
    tests := []struct {
        name string
        auth *configv1.BasicAuth
        expectErr string
    }{
        {
            name: "empty_username",
            auth: &configv1.BasicAuth{
                Username: proto.String(""),
            },
            expectErr: "username",
        },
         {
             name: "invalid_secret",
             auth: &configv1.BasicAuth{
                 Username: proto.String("user"),
                 Password: &configv1.SecretValue{Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MISSING_VAR"}},
             },
             expectErr: "basic auth password validation failed",
         },
         {
             name: "empty_secret_resolved",
             auth: &configv1.BasicAuth{
                 Username: proto.String("user"),
                 Password: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: ""}},
             },
             expectErr: "basic auth 'password' is empty",
         },
         {
             name: "valid_basic_auth",
             auth: &configv1.BasicAuth{
                 Username: proto.String("user"),
                 Password: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "pass"}},
             },
             expectErr: "",
         },
    }
     for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateBasicAuth(ctx, tt.auth)
            if tt.expectErr != "" {
                assert.ErrorContains(t, err, tt.expectErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
