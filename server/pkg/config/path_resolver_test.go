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
	baseDir := "/tmp/config"

	tests := []struct {
		name     string
		input    *configv1.McpAnyServerConfig
		expected *configv1.McpAnyServerConfig
	}{
		{
			name: "Stdio with empty WD",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
							McpService: &configv1.McpUpstreamService{
								ConnectionType: &configv1.McpUpstreamService_StdioConnection{
									StdioConnection: &configv1.McpStdioConnection{
										WorkingDirectory: proto.String(""),
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
										WorkingDirectory: proto.String(baseDir),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Stdio with relative WD",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
							McpService: &configv1.McpUpstreamService{
								ConnectionType: &configv1.McpUpstreamService_StdioConnection{
									StdioConnection: &configv1.McpStdioConnection{
										WorkingDirectory: proto.String("scripts"),
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
										WorkingDirectory: proto.String(filepath.Join(baseDir, "scripts")),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Stdio with absolute WD",
			input: &configv1.McpAnyServerConfig{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					{
						ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
							McpService: &configv1.McpUpstreamService{
								ConnectionType: &configv1.McpUpstreamService_StdioConnection{
									StdioConnection: &configv1.McpStdioConnection{
										WorkingDirectory: proto.String("/absolute/path"),
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
										WorkingDirectory: proto.String("/absolute/path"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Profile Secrets with relative path",
			input: &configv1.McpAnyServerConfig{
				GlobalSettings: &configv1.GlobalSettings{
					ProfileDefinitions: []*configv1.ProfileDefinition{
						{
							Secrets: map[string]*configv1.SecretValue{
								"my-secret": {
									Value: &configv1.SecretValue_FilePath{
										FilePath: "secrets/key.txt",
									},
								},
							},
						},
					},
				},
			},
			expected: &configv1.McpAnyServerConfig{
				GlobalSettings: &configv1.GlobalSettings{
					ProfileDefinitions: []*configv1.ProfileDefinition{
						{
							Secrets: map[string]*configv1.SecretValue{
								"my-secret": {
									Value: &configv1.SecretValue_FilePath{
										FilePath: filepath.Join(baseDir, "secrets/key.txt"),
									},
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

			if tt.expected.GlobalSettings != nil {
				if assert.NotNil(t, tt.input.GlobalSettings) {
					// Check secrets
					for k, v := range tt.expected.GlobalSettings.ProfileDefinitions[0].Secrets {
						actual := tt.input.GlobalSettings.ProfileDefinitions[0].Secrets[k]
						assert.Equal(t, v.GetFilePath(), actual.GetFilePath())
					}
				}
			}

			if len(tt.expected.UpstreamServices) > 0 {
				actualSvc := tt.input.UpstreamServices[0]
				expectedSvc := tt.expected.UpstreamServices[0]

				if expectedSvc.GetMcpService().GetStdioConnection() != nil {
					assert.Equal(t, expectedSvc.GetMcpService().GetStdioConnection().GetWorkingDirectory(), actualSvc.GetMcpService().GetStdioConnection().GetWorkingDirectory())
				}
			}
		})
	}
}
