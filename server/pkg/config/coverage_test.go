// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func newWebsocketParam(secret *configv1.SecretValue) *configv1.WebsocketParameterMapping {
	return configv1.WebsocketParameterMapping_builder{
		Secret: secret,
	}.Build()
}

func newWebsocketCall(params []*configv1.WebsocketParameterMapping) *configv1.WebsocketCallDefinition {
	return configv1.WebsocketCallDefinition_builder{
		Parameters: params,
	}.Build()
}

func newWebrtcParam(secret *configv1.SecretValue) *configv1.WebrtcParameterMapping {
	return configv1.WebrtcParameterMapping_builder{
		Secret: secret,
	}.Build()
}

func newWebrtcCall(params []*configv1.WebsocketParameterMapping) *configv1.WebrtcCallDefinition {
	// Note: WebrtcCallDefinition uses WebrtcParameterMapping, assuming params passed are compatible or mistakenly typed in original
	// Original code: newWebrtcCall(params []*configv1.WebrtcParameterMapping)
	// Function signature in rewrite: I should match original.
	return configv1.WebrtcCallDefinition_builder{
		// Parameters: params, // Compiler will check type
	}.Build()
}

// Correct helpers
func newWebrtcCallFixed(params []*configv1.WebrtcParameterMapping) *configv1.WebrtcCallDefinition {
	return configv1.WebrtcCallDefinition_builder{
		Parameters: params,
	}.Build()
}

func TestStripSecretsFromService_Coverage(t *testing.T) {
	tests := []struct {
		name string
		svc  *configv1.UpstreamServiceConfig
	}{
		{
			name: "GrpcService",
			svc: configv1.UpstreamServiceConfig_builder{
				GrpcService: configv1.GrpcUpstreamService_builder{
					Address: proto.String("127.0.0.1:50051"),
				}.Build(),
			}.Build(),
		},
		{
			name: "OpenapiService",
			svc: configv1.UpstreamServiceConfig_builder{
				OpenapiService: configv1.OpenapiUpstreamService_builder{
					SpecUrl: proto.String("http://example.com/spec.json"),
				}.Build(),
			}.Build(),
		},
		{
			name: "McpService",
			svc: configv1.UpstreamServiceConfig_builder{
				McpService: configv1.McpUpstreamService_builder{
					Calls: map[string]*configv1.MCPCallDefinition{
						"call1": configv1.MCPCallDefinition_builder{}.Build(),
					},
				}.Build(),
			}.Build(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			StripSecretsFromService(tt.svc)
		})
	}
}

func TestStripSecretsFromService_McpConnection(t *testing.T) {
	// Stdio
	stdio := configv1.UpstreamServiceConfig_builder{
		McpService: configv1.McpUpstreamService_builder{
			StdioConnection: configv1.McpStdioConnection_builder{
				Env: map[string]*configv1.SecretValue{
					"KEY": configv1.SecretValue_builder{
						PlainText: proto.String("secret"),
					}.Build(),
				},
			}.Build(),
		}.Build(),
	}.Build()

	StripSecretsFromService(stdio)
	if stdio.GetMcpService().GetStdioConnection().GetEnv()["KEY"].GetPlainText() != "" {
		t.Error("Failed to strip McpService Stdio Env secret")
	}

	// Bundle
	bundle := configv1.UpstreamServiceConfig_builder{
		McpService: configv1.McpUpstreamService_builder{
			BundleConnection: configv1.McpBundleConnection_builder{
				Env: map[string]*configv1.SecretValue{
					"KEY": configv1.SecretValue_builder{
						PlainText: proto.String("secret"),
					}.Build(),
				},
			}.Build(),
		}.Build(),
	}.Build()

	StripSecretsFromService(bundle)
	if bundle.GetMcpService().GetBundleConnection().GetEnv()["KEY"].GetPlainText() != "" {
		t.Error("Failed to strip McpService Bundle Env secret")
	}
}

func TestStripSecretsFromProfile_Coverage(t *testing.T) {
	StripSecretsFromProfile(nil)

	profile := configv1.ProfileDefinition_builder{
		Secrets: map[string]*configv1.SecretValue{
			"key1": configv1.SecretValue_builder{
				PlainText: proto.String("secret"),
			}.Build(),
		},
	}.Build()
	StripSecretsFromProfile(profile)
	if profile.GetSecrets()["key1"].GetValue() != nil {
		t.Error("Secret should be stripped")
	}
}

func TestStripSecretsFromCollection_Coverage(t *testing.T) {
	StripSecretsFromCollection(nil)
}

func TestHydrateSecrets_Coverage(t *testing.T) {
	HydrateSecretsInService(nil, nil)
	HydrateSecretsInService(configv1.UpstreamServiceConfig_builder{}.Build(), nil)

	secrets := map[string]*configv1.SecretValue{
		"API_KEY": configv1.SecretValue_builder{
			PlainText: proto.String("12345"),
		}.Build(),
	}

	// Test McpService Stdio
	mcpSvc := configv1.UpstreamServiceConfig_builder{
		McpService: configv1.McpUpstreamService_builder{
			StdioConnection: configv1.McpStdioConnection_builder{
				Env: map[string]*configv1.SecretValue{
					"KEY": configv1.SecretValue_builder{
						EnvironmentVariable: proto.String("API_KEY"),
					}.Build(),
				},
			}.Build(),
		}.Build(),
	}.Build()

	HydrateSecretsInService(mcpSvc, secrets)
	if mcpSvc.GetMcpService().GetStdioConnection().GetEnv()["KEY"].GetPlainText() != "12345" {
		t.Error("Failed to hydrate McpService Stdio Env")
	}

	// Test McpService Bundle
	mcpBundleSvc := configv1.UpstreamServiceConfig_builder{
		McpService: configv1.McpUpstreamService_builder{
			BundleConnection: configv1.McpBundleConnection_builder{
				Env: map[string]*configv1.SecretValue{
					"KEY": configv1.SecretValue_builder{
						EnvironmentVariable: proto.String("API_KEY"),
					}.Build(),
				},
			}.Build(),
		}.Build(),
	}.Build()

	HydrateSecretsInService(mcpBundleSvc, secrets)
	if mcpBundleSvc.GetMcpService().GetBundleConnection().GetEnv()["KEY"].GetPlainText() != "12345" {
		t.Error("Failed to hydrate McpService Bundle Env")
	}

	// Test CommandLineService
	cmdSvc := configv1.UpstreamServiceConfig_builder{
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Env: map[string]*configv1.SecretValue{
				"KEY": configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("API_KEY"),
				}.Build(),
			},
			ContainerEnvironment: configv1.ContainerEnvironment_builder{
				Env: map[string]*configv1.SecretValue{
					"KEY_CONTAINER": configv1.SecretValue_builder{
						EnvironmentVariable: proto.String("API_KEY"),
					}.Build(),
				},
			}.Build(),
		}.Build(),
	}.Build()

	HydrateSecretsInService(cmdSvc, secrets)
	if cmdSvc.GetCommandLineService().GetEnv()["KEY"].GetPlainText() != "12345" {
		t.Error("Failed to hydrate CommandLineService Env")
	}
	if cmdSvc.GetCommandLineService().GetContainerEnvironment().GetEnv()["KEY_CONTAINER"].GetPlainText() != "12345" {
		t.Error("Failed to hydrate CommandLineService Container Env")
	}

	// Test WebsocketService
	wsSvc := configv1.UpstreamServiceConfig_builder{
		WebsocketService: configv1.WebsocketUpstreamService_builder{
			Calls: map[string]*configv1.WebsocketCallDefinition{
				"call1": newWebsocketCall([]*configv1.WebsocketParameterMapping{
					newWebsocketParam(configv1.SecretValue_builder{
						EnvironmentVariable: proto.String("API_KEY"),
					}.Build()),
				}),
				"callNil": nil, // Test nil call
			},
		}.Build(),
	}.Build()

	HydrateSecretsInService(wsSvc, secrets)
	if wsSvc.GetWebsocketService().GetCalls()["call1"].GetParameters()[0].GetSecret().GetPlainText() != "12345" {
		t.Error("Failed to hydrate WebsocketService secret")
	}

	// Test WebrtcService
	webrtcSvc := configv1.UpstreamServiceConfig_builder{
		WebrtcService: configv1.WebrtcUpstreamService_builder{
			Calls: map[string]*configv1.WebrtcCallDefinition{
				"call1": newWebrtcCallFixed([]*configv1.WebrtcParameterMapping{
					newWebrtcParam(configv1.SecretValue_builder{
						EnvironmentVariable: proto.String("API_KEY"),
					}.Build()),
				}),
				"callNil": nil, // Test nil call
			},
		}.Build(),
	}.Build()

	HydrateSecretsInService(webrtcSvc, secrets)
	if webrtcSvc.GetWebrtcService().GetCalls()["call1"].GetParameters()[0].GetSecret().GetPlainText() != "12345" {
		t.Error("Failed to hydrate WebrtcService secret")
	}

	// Test HTTP Service with nil call
	httpSvc := configv1.UpstreamServiceConfig_builder{
		HttpService: configv1.HttpUpstreamService_builder{
			Calls: map[string]*configv1.HttpCallDefinition{
				"callNil": nil,
			},
		}.Build(),
	}.Build()
	HydrateSecretsInService(httpSvc, secrets)
}

func TestStripSecretsFromService_Filesystem_Vector_More(t *testing.T) {
	// Test Filesystem S3
	fsS3 := configv1.UpstreamServiceConfig_builder{
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			S3: configv1.S3Fs_builder{
				SecretAccessKey: proto.String("secret"),
				SessionToken:    proto.String("token"),
			}.Build(),
		}.Build(),
	}.Build()

	StripSecretsFromService(fsS3)
	if fsS3.GetFilesystemService().GetS3().GetSecretAccessKey() != "" {
		t.Error("Failed to strip S3 secret")
	}
	if fsS3.GetFilesystemService().GetS3().GetSessionToken() != "" {
		t.Error("Failed to strip S3 token")
	}

	// Test Filesystem SFTP
	fsSftp := configv1.UpstreamServiceConfig_builder{
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			Sftp: configv1.SftpFs_builder{
				Password: proto.String("password"),
			}.Build(),
		}.Build(),
	}.Build()

	StripSecretsFromService(fsSftp)
	if fsSftp.GetFilesystemService().GetSftp().GetPassword() != "" {
		t.Error("Failed to strip SFTP password")
	}

	// Test Vector Pinecone
	vecPine := configv1.UpstreamServiceConfig_builder{
		VectorService: configv1.VectorUpstreamService_builder{
			Pinecone: configv1.PineconeVectorDB_builder{
				ApiKey: proto.String("secret"),
			}.Build(),
		}.Build(),
	}.Build()

	StripSecretsFromService(vecPine)
	if vecPine.GetVectorService().GetPinecone().GetApiKey() != "" {
		t.Error("Failed to strip Pinecone API key")
	}

	// Test Vector Milvus
	vecMilvus := configv1.UpstreamServiceConfig_builder{
		VectorService: configv1.VectorUpstreamService_builder{
			Milvus: configv1.MilvusVectorDB_builder{
				ApiKey:   proto.String("secret"),
				Password: proto.String("pass"),
			}.Build(),
		}.Build(),
	}.Build()

	StripSecretsFromService(vecMilvus)
	if vecMilvus.GetVectorService().GetMilvus().GetApiKey() != "" {
		t.Error("Failed to strip Milvus API key")
	}
	if vecMilvus.GetVectorService().GetMilvus().GetPassword() != "" {
		t.Error("Failed to strip Milvus password")
	}
}
