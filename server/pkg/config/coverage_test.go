package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestStripSecretsFromService_Coverage(t *testing.T) {
	tests := []struct {
		name string
		svc  *configv1.UpstreamServiceConfig
	}{
		{
			name: "GrpcService",
			svc: func() *configv1.UpstreamServiceConfig {
				grpcSvc := configv1.GrpcUpstreamService_builder{
					Address: proto.String("127.0.0.1:50051"),
				}.Build()

				return configv1.UpstreamServiceConfig_builder{
					GrpcService: grpcSvc,
				}.Build()
			}(),
		},
		{
			name: "OpenapiService",
			svc: func() *configv1.UpstreamServiceConfig {
				openapiSvc := configv1.OpenapiUpstreamService_builder{
					SpecUrl: proto.String("http://example.com/spec.json"),
				}.Build()

				return configv1.UpstreamServiceConfig_builder{
					OpenapiService: openapiSvc,
				}.Build()
			}(),
		},
		{
			name: "McpService",
			svc: func() *configv1.UpstreamServiceConfig {
				mcpSvc := configv1.McpUpstreamService_builder{
					Calls: map[string]*configv1.MCPCallDefinition{
						"call1": configv1.MCPCallDefinition_builder{}.Build(),
					},
				}.Build()

				return configv1.UpstreamServiceConfig_builder{
					McpService: mcpSvc,
				}.Build()
			}(),
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
	stdio := configv1.UpstreamServiceConfig_builder{}.Build()
	{
		conn := configv1.McpStdioConnection_builder{
			Env: map[string]*configv1.SecretValue{
				"KEY": configv1.SecretValue_builder{
					PlainText: proto.String("secret"),
				}.Build(),
			},
		}.Build()

		mcpSvc := configv1.McpUpstreamService_builder{
			StdioConnection: conn,
		}.Build()
		stdio.SetMcpService(mcpSvc) // Setters are still fine/required if we are modifying existing object or if we want to mix.
		// But better to rebuild if we want fully pure opaque usage, but here we just need initial state.
		// Wait, `stdio` is created empty.
		// The test logic seems to construct incrementally.
		// I will respect the structure but use builders for creation.
	}
	StripSecretsFromService(stdio)
	if stdio.GetMcpService().GetStdioConnection().GetEnv()["KEY"].GetPlainText() != "" {
		t.Error("Failed to strip McpService Stdio Env secret")
	}

	// Bundle
	bundle := configv1.UpstreamServiceConfig_builder{}.Build()
	{
		conn := configv1.McpBundleConnection_builder{
			Env: map[string]*configv1.SecretValue{
				"KEY": configv1.SecretValue_builder{
					PlainText: proto.String("secret"),
				}.Build(),
			},
		}.Build()

		mcpSvc := configv1.McpUpstreamService_builder{
			BundleConnection: conn,
		}.Build()
		bundle.SetMcpService(mcpSvc)
	}
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
	if profile.GetSecrets()["key1"].GetPlainText() != "" {
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
	mcpSvc := configv1.UpstreamServiceConfig_builder{}.Build()
	{
		conn := configv1.McpStdioConnection_builder{
			Env: map[string]*configv1.SecretValue{
				"KEY": configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("API_KEY"),
				}.Build(),
			},
		}.Build()

		svc := configv1.McpUpstreamService_builder{
			StdioConnection: conn,
		}.Build()
		mcpSvc.SetMcpService(svc)
	}
	HydrateSecretsInService(mcpSvc, secrets)
	if mcpSvc.GetMcpService().GetStdioConnection().GetEnv()["KEY"].GetPlainText() != "12345" {
		t.Error("Failed to hydrate McpService Stdio Env")
	}

	// Test McpService Bundle
	mcpBundleSvc := configv1.UpstreamServiceConfig_builder{}.Build()
	{
		conn := configv1.McpBundleConnection_builder{
			Env: map[string]*configv1.SecretValue{
				"KEY": configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("API_KEY"),
				}.Build(),
			},
		}.Build()

		svc := configv1.McpUpstreamService_builder{
			BundleConnection: conn,
		}.Build()
		mcpBundleSvc.SetMcpService(svc)
	}
	HydrateSecretsInService(mcpBundleSvc, secrets)
	if mcpBundleSvc.GetMcpService().GetBundleConnection().GetEnv()["KEY"].GetPlainText() != "12345" {
		t.Error("Failed to hydrate McpService Bundle Env")
	}

	// Test CommandLineService
	cmdSvc := configv1.UpstreamServiceConfig_builder{}.Build()
	{
		svc := configv1.CommandLineUpstreamService_builder{
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
		}.Build()

		cmdSvc.SetCommandLineService(svc)
	}
	HydrateSecretsInService(cmdSvc, secrets)
	if cmdSvc.GetCommandLineService().GetEnv()["KEY"].GetPlainText() != "12345" {
		t.Error("Failed to hydrate CommandLineService Env")
	}
	if cmdSvc.GetCommandLineService().GetContainerEnvironment().GetEnv()["KEY_CONTAINER"].GetPlainText() != "12345" {
		t.Error("Failed to hydrate CommandLineService Container Env")
	}

	// Test WebsocketService
	wsSvc := configv1.UpstreamServiceConfig_builder{}.Build()
	{
		paramMapping := configv1.WebsocketParameterMapping_builder{
			Secret: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("API_KEY"),
			}.Build(),
		}.Build()

		callDef := configv1.WebsocketCallDefinition_builder{
			Parameters: []*configv1.WebsocketParameterMapping{paramMapping},
		}.Build()

		svc := configv1.WebsocketUpstreamService_builder{
			Calls: map[string]*configv1.WebsocketCallDefinition{
				"call1":   callDef,
				"callNil": nil,
			},
		}.Build()

		wsSvc.SetWebsocketService(svc)
	}
	HydrateSecretsInService(wsSvc, secrets)
	if wsSvc.GetWebsocketService().GetCalls()["call1"].GetParameters()[0].GetSecret().GetPlainText() != "12345" {
		t.Error("Failed to hydrate WebsocketService secret")
	}

	// Test WebrtcService
	webrtcSvc := configv1.UpstreamServiceConfig_builder{}.Build()
	{
		paramMapping := configv1.WebrtcParameterMapping_builder{
			Secret: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("API_KEY"),
			}.Build(),
		}.Build()

		callDef := configv1.WebrtcCallDefinition_builder{
			Parameters: []*configv1.WebrtcParameterMapping{paramMapping},
		}.Build()

		svc := configv1.WebrtcUpstreamService_builder{
			Calls: map[string]*configv1.WebrtcCallDefinition{
				"call1":   callDef,
				"callNil": nil,
			},
		}.Build()

		webrtcSvc.SetWebrtcService(svc)
	}
	HydrateSecretsInService(webrtcSvc, secrets)
	if webrtcSvc.GetWebrtcService().GetCalls()["call1"].GetParameters()[0].GetSecret().GetPlainText() != "12345" {
		t.Error("Failed to hydrate WebrtcService secret")
	}

	// Test HTTP Service with nil call
	httpSvc := configv1.UpstreamServiceConfig_builder{}.Build()
	{
		svc := configv1.HttpUpstreamService_builder{
			Calls: map[string]*configv1.HttpCallDefinition{
				"callNil": nil,
			},
		}.Build()
		httpSvc.SetHttpService(svc)
	}
	HydrateSecretsInService(httpSvc, secrets)
}

func TestStripSecretsFromService_Filesystem_Vector_More(t *testing.T) {
	// Test Filesystem S3
	fsS3 := configv1.UpstreamServiceConfig_builder{}.Build()
	{
		s3 := configv1.S3Fs_builder{
			SecretAccessKey: proto.String("secret"),
			SessionToken:    proto.String("token"),
		}.Build()

		svc := configv1.FilesystemUpstreamService_builder{
			S3: s3,
		}.Build()
		fsS3.SetFilesystemService(svc)
	}
	StripSecretsFromService(fsS3)
	if fsS3.GetFilesystemService().GetS3().GetSecretAccessKey() != "" {
		t.Error("Failed to strip S3 secret")
	}
	if fsS3.GetFilesystemService().GetS3().GetSessionToken() != "" {
		t.Error("Failed to strip S3 token")
	}

	// Test Filesystem SFTP
	fsSftp := configv1.UpstreamServiceConfig_builder{}.Build()
	{
		sftp := configv1.SftpFs_builder{
			Password: proto.String("password"),
		}.Build()

		svc := configv1.FilesystemUpstreamService_builder{
			Sftp: sftp,
		}.Build()
		fsSftp.SetFilesystemService(svc)
	}
	StripSecretsFromService(fsSftp)
	if fsSftp.GetFilesystemService().GetSftp().GetPassword() != "" {
		t.Error("Failed to strip SFTP password")
	}

	// Test Vector Pinecone
	vecPine := configv1.UpstreamServiceConfig_builder{}.Build()
	{
		pinecone := configv1.PineconeVectorDB_builder{
			ApiKey: proto.String("secret"),
		}.Build()

		svc := configv1.VectorUpstreamService_builder{
			Pinecone: pinecone,
		}.Build()
		vecPine.SetVectorService(svc)
	}

	StripSecretsFromService(vecPine)
	if vecPine.GetVectorService().GetPinecone().GetApiKey() != "" {
		t.Error("Failed to strip Pinecone API key")
	}

	// Test Vector Milvus
	vecMilvus := configv1.UpstreamServiceConfig_builder{}.Build()
	{
		milvus := configv1.MilvusVectorDB_builder{
			ApiKey:   proto.String("secret"),
			Password: proto.String("pass"),
		}.Build()

		svc := configv1.VectorUpstreamService_builder{
			Milvus: milvus,
		}.Build()
		vecMilvus.SetVectorService(svc)
	}

	StripSecretsFromService(vecMilvus)
	if vecMilvus.GetVectorService().GetMilvus().GetApiKey() != "" {
		t.Error("Failed to strip Milvus API key")
	}
	if vecMilvus.GetVectorService().GetMilvus().GetPassword() != "" {
		t.Error("Failed to strip Milvus password")
	}
}
