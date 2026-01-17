// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestResolveRelativePaths(t *testing.T) {
	baseDir := "/base/dir"

	tests := []struct {
		name     string
		input    *configv1.McpAnyServerConfig
		expected *configv1.McpAnyServerConfig
	}{
		{
			name: "Resolve CommandLine Service",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
							CommandLineService: &configv1.CommandLineUpstreamService{
								Command:          stringPtr("./script.sh"),
								WorkingDirectory: stringPtr("workdir"),
							},
						},
					},
				},
			},
			expected: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
							CommandLineService: &configv1.CommandLineUpstreamService{
								Command:          stringPtr(filepath.Join(baseDir, "./script.sh")),
								WorkingDirectory: stringPtr(filepath.Join(baseDir, "workdir")),
							},
						},
					},
				},
			},
		},
		{
			name: "Resolve Mcp Stdio Service",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
							McpService: &configv1.McpUpstreamService{
								ConnectionType: &configv1.McpUpstreamService_StdioConnection{
									StdioConnection: &configv1.McpStdioConnection{
										Command:          stringPtr("node"), // Should not change
										Args:             []string{"./index.js", "--flag"},
										WorkingDirectory: stringPtr("./app"),
									},
								},
							},
						},
					},
				},
			},
			expected: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
							McpService: &configv1.McpUpstreamService{
								ConnectionType: &configv1.McpUpstreamService_StdioConnection{
									StdioConnection: &configv1.McpStdioConnection{
										Command:          stringPtr("node"),
										Args:             []string{filepath.Join(baseDir, "./index.js"), "--flag"},
										WorkingDirectory: stringPtr(filepath.Join(baseDir, "./app")),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Resolve Mcp Stdio Service - Absolute Path",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
							McpService: &configv1.McpUpstreamService{
								ConnectionType: &configv1.McpUpstreamService_StdioConnection{
									StdioConnection: &configv1.McpStdioConnection{
										Command:          stringPtr("/usr/bin/python"),
										Args:             []string{"/opt/script.py"},
										WorkingDirectory: stringPtr("/tmp"),
									},
								},
							},
						},
					},
				},
			},
			expected: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
							McpService: &configv1.McpUpstreamService{
								ConnectionType: &configv1.McpUpstreamService_StdioConnection{
									StdioConnection: &configv1.McpStdioConnection{
										Command:          stringPtr("/usr/bin/python"),
										Args:             []string{"/opt/script.py"},
										WorkingDirectory: stringPtr("/tmp"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Resolve Secret File Path",
			input: &configv1.McpAnyServerConfig{
				GlobalSettings: &configv1.GlobalSettings{
					ApiKey: stringPtr("somekey"),
				},
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						UpstreamAuth: &configv1.Authentication{
							AuthMethod: &configv1.Authentication_ApiKey{
								ApiKey: &configv1.APIKeyAuth{
									Value: &configv1.SecretValue{
										Value: &configv1.SecretValue_FilePath{
											FilePath: "secret.txt",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: &configv1.McpAnyServerConfig{
				GlobalSettings: &configv1.GlobalSettings{
					ApiKey: stringPtr("somekey"),
				},
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						UpstreamAuth: &configv1.Authentication{
							AuthMethod: &configv1.Authentication_ApiKey{
								ApiKey: &configv1.APIKeyAuth{
									Value: &configv1.SecretValue{
										Value: &configv1.SecretValue_FilePath{
											FilePath: filepath.Join(baseDir, "secret.txt"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Resolve OpenAPI Spec URL (File Path)",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
							OpenapiService: &configv1.OpenapiUpstreamService{
								SpecSource: &configv1.OpenapiUpstreamService_SpecUrl{
									SpecUrl: "openapi.yaml",
								},
							},
						},
					},
				},
			},
			expected: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
							OpenapiService: &configv1.OpenapiUpstreamService{
								SpecSource: &configv1.OpenapiUpstreamService_SpecUrl{
									SpecUrl: filepath.Join(baseDir, "openapi.yaml"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Resolve OpenAPI Spec URL (HTTP)",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
							OpenapiService: &configv1.OpenapiUpstreamService{
								SpecSource: &configv1.OpenapiUpstreamService_SpecUrl{
									SpecUrl: "http://example.com/openapi.yaml",
								},
							},
						},
					},
				},
			},
			expected: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
							OpenapiService: &configv1.OpenapiUpstreamService{
								SpecSource: &configv1.OpenapiUpstreamService_SpecUrl{
									SpecUrl: "http://example.com/openapi.yaml",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResolveRelativePaths(tt.input, baseDir)
			assert.True(t, proto.Equal(tt.input, tt.expected), "expected: %v, got: %v", tt.expected, tt.input)
		})
	}
}
