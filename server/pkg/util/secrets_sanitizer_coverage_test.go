// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestStripSecretsFromService_Coverage(t *testing.T) {
	// CommandLineService via StripSecretsFromService
	cmdSvc := &configv1.UpstreamServiceConfig{
		Name: proto.String("cmd"),
		ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
			CommandLineService: &configv1.CommandLineUpstreamService{
				Env: map[string]*configv1.SecretValue{
					"K": {Value: &configv1.SecretValue_PlainText{PlainText: "s"}},
				},
				Calls: map[string]*configv1.CommandLineCallDefinition{
					"c": {Parameters: []*configv1.CommandLineParameterMapping{{Secret: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "s"}}}}},
				},
			},
		},
	}
	StripSecretsFromService(cmdSvc)
	assert.Nil(t, cmdSvc.GetCommandLineService().Env["K"].Value)
	assert.Nil(t, cmdSvc.GetCommandLineService().Calls["c"].Parameters[0].Secret.Value)

	// McpService via StripSecretsFromService (Stdio)
	mcpSvc := &configv1.UpstreamServiceConfig{
		Name: proto.String("mcp"),
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_StdioConnection{
					StdioConnection: &configv1.McpStdioConnection{
						Env: map[string]*configv1.SecretValue{"K": {Value: &configv1.SecretValue_PlainText{PlainText: "s"}}},
					},
				},
				Calls: map[string]*configv1.MCPCallDefinition{
					"c": {InputSchema: nil},
				},
			},
		},
	}
	StripSecretsFromService(mcpSvc)
	assert.Nil(t, mcpSvc.GetMcpService().GetStdioConnection().Env["K"].Value)

	// McpService Bundle
	mcpBundleSvc := &configv1.UpstreamServiceConfig{
		Name: proto.String("mcp-bundle"),
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_BundleConnection{
					BundleConnection: &configv1.McpBundleConnection{
						Env: map[string]*configv1.SecretValue{"K": {Value: &configv1.SecretValue_PlainText{PlainText: "s"}}},
					},
				},
			},
		},
	}
	StripSecretsFromService(mcpBundleSvc)
	assert.Nil(t, mcpBundleSvc.GetMcpService().GetBundleConnection().Env["K"].Value)

	// Vector Milvus
	vecSvc := &configv1.UpstreamServiceConfig{
		Name: proto.String("vector"),
		ServiceConfig: &configv1.UpstreamServiceConfig_VectorService{
			VectorService: &configv1.VectorUpstreamService{
				VectorDbType: &configv1.VectorUpstreamService_Milvus{
					Milvus: &configv1.MilvusVectorDB{
						ApiKey: proto.String("s"),
						Password: proto.String("p"),
					},
				},
			},
		},
	}
	StripSecretsFromService(vecSvc)
	assert.Equal(t, "", vecSvc.GetVectorService().GetMilvus().GetApiKey())
	assert.Equal(t, "", vecSvc.GetVectorService().GetMilvus().GetPassword())

	// Filesystem SFTP
	fsSvc := &configv1.UpstreamServiceConfig{
		Name: proto.String("fs"),
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Sftp{
					Sftp: &configv1.SftpFs{
						Password: proto.String("p"),
					},
				},
			},
		},
	}
	StripSecretsFromService(fsSvc)
	assert.Equal(t, "", fsSvc.GetFilesystemService().GetSftp().GetPassword())

	// Graphql and Sql (empty cases)
	gqlSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_GraphqlService{GraphqlService: &configv1.GraphQLUpstreamService{}},
	}
	StripSecretsFromService(gqlSvc) // Should not panic

	sqlSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_SqlService{SqlService: &configv1.SqlUpstreamService{}},
	}
	StripSecretsFromService(sqlSvc) // Should not panic
}

func TestHydrateSecretsInService_Coverage(t *testing.T) {
	secrets := map[string]*configv1.SecretValue{
		"K": {Value: &configv1.SecretValue_PlainText{PlainText: "s"}},
	}

	// CommandLine Service
	cmdSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
			CommandLineService: &configv1.CommandLineUpstreamService{
				Env: map[string]*configv1.SecretValue{
					"E": {Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "K"}},
				},
				ContainerEnvironment: &configv1.ContainerEnvironment{
					Env: map[string]*configv1.SecretValue{
						"C": {Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "K"}},
					},
				},
			},
		},
	}
	HydrateSecretsInService(cmdSvc, secrets)
	assert.Equal(t, "s", cmdSvc.GetCommandLineService().Env["E"].GetPlainText())
	assert.Equal(t, "s", cmdSvc.GetCommandLineService().ContainerEnvironment.Env["C"].GetPlainText())

	// McpService Stdio and Bundle
	mcpStdio := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_StdioConnection{
					StdioConnection: &configv1.McpStdioConnection{
						Env: map[string]*configv1.SecretValue{"E": {Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "K"}}},
					},
				},
			},
		},
	}
	HydrateSecretsInService(mcpStdio, secrets)
	assert.Equal(t, "s", mcpStdio.GetMcpService().GetStdioConnection().Env["E"].GetPlainText())

	mcpBundle := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_BundleConnection{
					BundleConnection: &configv1.McpBundleConnection{
						Env: map[string]*configv1.SecretValue{"E": {Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "K"}}},
					},
				},
			},
		},
	}
	HydrateSecretsInService(mcpBundle, secrets)
	assert.Equal(t, "s", mcpBundle.GetMcpService().GetBundleConnection().Env["E"].GetPlainText())
}

func TestNilChecks_Coverage(t *testing.T) {
	StripSecretsFromService(nil)
	StripSecretsFromProfile(nil)
	StripSecretsFromCollection(nil)
	StripSecretsFromAuth(nil)
	HydrateSecretsInService(nil, nil)

	// Hydrate with empty secrets
	svc := &configv1.UpstreamServiceConfig{}
	HydrateSecretsInService(svc, nil)

	// Nil sub-configs
	stripSecretsFromCommandLineService(nil)
	stripSecretsFromHTTPService(nil)
	stripSecretsFromMcpService(nil)
	stripSecretsFromFilesystemService(nil)
	stripSecretsFromVectorService(nil)
	stripSecretsFromWebsocketService(nil)
	stripSecretsFromWebrtcService(nil)
	stripSecretsFromHook(nil)
	stripSecretsFromCacheConfig(nil)

	hydrateSecretsInHTTPService(nil, nil)
	hydrateSecretsInWebsocketService(nil, nil)
	hydrateSecretsInWebrtcService(nil, nil)

	// Calls with nil
	stripSecretsFromCommandLineCall(nil)
	stripSecretsFromHTTPCall(nil)
	stripSecretsFromWebsocketCall(nil)
	stripSecretsFromWebrtcCall(nil)

	// Hydrate calls with nil
	hydrateSecretsInHTTPService(&configv1.HttpUpstreamService{Calls: map[string]*configv1.HttpCallDefinition{"c": nil}}, nil)
	hydrateSecretsInWebsocketService(&configv1.WebsocketUpstreamService{Calls: map[string]*configv1.WebsocketCallDefinition{"c": nil}}, nil)
	hydrateSecretsInWebrtcService(&configv1.WebrtcUpstreamService{Calls: map[string]*configv1.WebrtcCallDefinition{"c": nil}}, nil)
}
