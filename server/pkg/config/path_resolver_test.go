// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestResolveRelativePaths(t *testing.T) {
	baseDir := "/abs/path/to/config"

	tests := []struct {
		name     string
		input    *configv1.McpAnyServerConfig
		expected *configv1.McpAnyServerConfig
	}{
		{
			name: "Command Line Service - Relative Paths",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
							CommandLineService: &configv1.CommandLineUpstreamService{
								Command:          proto.String("./script.sh"),
								WorkingDirectory: proto.String("subdir"),
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
								Command:          proto.String(filepath.Join(baseDir, "./script.sh")),
								WorkingDirectory: proto.String(filepath.Join(baseDir, "subdir")),
							},
						},
					},
				},
			},
		},
		{
			name: "Command Line Service - System Command (No Resolution)",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
							CommandLineService: &configv1.CommandLineUpstreamService{
								Command:          proto.String("npm"),
								WorkingDirectory: proto.String(""),
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
								Command:          proto.String("npm"),
								WorkingDirectory: proto.String(""),
							},
						},
					},
				},
			},
		},
		{
			name: "Command Line Service - Absolute Paths (No Resolution)",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
							CommandLineService: &configv1.CommandLineUpstreamService{
								Command:          proto.String("/usr/bin/python"),
								WorkingDirectory: proto.String("/tmp"),
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
								Command:          proto.String("/usr/bin/python"),
								WorkingDirectory: proto.String("/tmp"),
							},
						},
					},
				},
			},
		},
		{
			name: "MCP Service - Stdio Connection",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
							McpService: &configv1.McpUpstreamService{
								ConnectionType: &configv1.McpUpstreamService_StdioConnection{
									StdioConnection: &configv1.McpStdioConnection{
										Command:          proto.String("python"), // System command
										WorkingDirectory: proto.String("./py_env"),
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
										Command:          proto.String("python"),
										WorkingDirectory: proto.String(filepath.Join(baseDir, "./py_env")),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Filesystem Service",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
							FilesystemService: &configv1.FilesystemUpstreamService{
								RootPaths: map[string]string{
									"/data": "./data",
								},
							},
						},
					},
				},
			},
			expected: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
							FilesystemService: &configv1.FilesystemUpstreamService{
								RootPaths: map[string]string{
									"/data": filepath.Join(baseDir, "./data"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "TLS Config",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
							HttpService: &configv1.HttpUpstreamService{
								TlsConfig: &configv1.TLSConfig{
									CaCertPath:     proto.String("certs/ca.pem"),
									ClientCertPath: proto.String("/etc/certs/client.pem"),
								},
							},
						},
					},
				},
			},
			expected: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
							HttpService: &configv1.HttpUpstreamService{
								TlsConfig: &configv1.TLSConfig{
									CaCertPath:     proto.String(filepath.Join(baseDir, "certs/ca.pem")),
									ClientCertPath: proto.String("/etc/certs/client.pem"),
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
			input := proto.Clone(tt.input).(*configv1.McpAnyServerConfig)
			ResolveRelativePaths(input, baseDir)

			if !proto.Equal(input, tt.expected) {
				t.Errorf("Expected:\n%v\nGot:\n%v", tt.expected, input)
			}
		})
	}
}
