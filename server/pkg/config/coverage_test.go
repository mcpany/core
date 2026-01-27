// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

func TestStripSecretsFromService_Coverage(t *testing.T) {
	tests := []struct {
		name string
		svc  *configv1.UpstreamServiceConfig
	}{
		{
			name: "GrpcService",
			svc: func() *configv1.UpstreamServiceConfig {
				grpcSvc := &configv1.GrpcUpstreamService{}
				grpcSvc.SetAddress("127.0.0.1:50051")

				cfg := &configv1.UpstreamServiceConfig{}
				cfg.SetGrpcService(grpcSvc)
				return cfg
			}(),
		},
		{
			name: "OpenapiService",
			svc: func() *configv1.UpstreamServiceConfig {
				openapiSvc := &configv1.OpenapiUpstreamService{}
				openapiSvc.SetSpecUrl("http://example.com/spec.json")

				cfg := &configv1.UpstreamServiceConfig{}
				cfg.SetOpenapiService(openapiSvc)
				return cfg
			}(),
		},
		{
			name: "McpService",
			svc: func() *configv1.UpstreamServiceConfig {
				mcpSvc := &configv1.McpUpstreamService{}
				mcpSvc.SetCalls(map[string]*configv1.MCPCallDefinition{
					"call1": {},
				})

				cfg := &configv1.UpstreamServiceConfig{}
				cfg.SetMcpService(mcpSvc)
				return cfg
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
    stdio := &configv1.UpstreamServiceConfig{}
	{
		conn := &configv1.McpStdioConnection{}
		conn.SetEnv(map[string]*configv1.SecretValue{
			"KEY": {
				Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
			},
		})

		mcpSvc := &configv1.McpUpstreamService{}
		mcpSvc.SetStdioConnection(conn)
		stdio.SetMcpService(mcpSvc)
	}
    StripSecretsFromService(stdio)
    if stdio.GetMcpService().GetStdioConnection().GetEnv()["KEY"].GetPlainText() != "" {
        t.Error("Failed to strip McpService Stdio Env secret")
    }

    // Bundle
    bundle := &configv1.UpstreamServiceConfig{}
	{
		conn := &configv1.McpBundleConnection{}
		conn.SetEnv(map[string]*configv1.SecretValue{
			"KEY": {
				Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
			},
		})

		mcpSvc := &configv1.McpUpstreamService{}
		mcpSvc.SetBundleConnection(conn)
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
			"key1": {
				Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
			},
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
	HydrateSecretsInService(&configv1.UpstreamServiceConfig{}, nil)

    secrets := map[string]*configv1.SecretValue{
        "API_KEY": {
            Value: &configv1.SecretValue_PlainText{PlainText: "12345"},
        },
    }

    // Test McpService Stdio
    mcpSvc := &configv1.UpstreamServiceConfig{}
	{
		conn := &configv1.McpStdioConnection{}
		conn.SetEnv(map[string]*configv1.SecretValue{
			"KEY": {
				Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "API_KEY"},
			},
		})

		svc := &configv1.McpUpstreamService{}
		svc.SetStdioConnection(conn)
		mcpSvc.SetMcpService(svc)
	}
    HydrateSecretsInService(mcpSvc, secrets)
    if mcpSvc.GetMcpService().GetStdioConnection().GetEnv()["KEY"].GetPlainText() != "12345" {
        t.Error("Failed to hydrate McpService Stdio Env")
    }

    // Test McpService Bundle
    mcpBundleSvc := &configv1.UpstreamServiceConfig{}
	{
		conn := &configv1.McpBundleConnection{}
		conn.SetEnv(map[string]*configv1.SecretValue{
			"KEY": {
				Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "API_KEY"},
			},
		})

		svc := &configv1.McpUpstreamService{}
		svc.SetBundleConnection(conn)
		mcpBundleSvc.SetMcpService(svc)
	}
    HydrateSecretsInService(mcpBundleSvc, secrets)
    if mcpBundleSvc.GetMcpService().GetBundleConnection().GetEnv()["KEY"].GetPlainText() != "12345" {
        t.Error("Failed to hydrate McpService Bundle Env")
    }

    // Test CommandLineService
    cmdSvc := &configv1.UpstreamServiceConfig{}
	{
		svc := &configv1.CommandLineUpstreamService{}
		svc.SetEnv(map[string]*configv1.SecretValue{
			"KEY": {
				Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "API_KEY"},
			},
		})

		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetEnv(map[string]*configv1.SecretValue{
			"KEY_CONTAINER": {
				Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "API_KEY"},
			},
		})
		svc.SetContainerEnvironment(containerEnv)

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
    wsSvc := &configv1.UpstreamServiceConfig{}
	{
		svc := &configv1.WebsocketUpstreamService{}

		paramMapping := &configv1.WebsocketParameterMapping{}
		secretVal := &configv1.SecretValue{}
		secretVal.SetEnvironmentVariable("API_KEY")
		paramMapping.SetSecret(secretVal)

		callDef := &configv1.WebsocketCallDefinition{}
		callDef.SetParameters([]*configv1.WebsocketParameterMapping{paramMapping})

		svc.SetCalls(map[string]*configv1.WebsocketCallDefinition{
			"call1": callDef,
			"callNil": nil,
		})

		wsSvc.SetWebsocketService(svc)
	}
    HydrateSecretsInService(wsSvc, secrets)
    if wsSvc.GetWebsocketService().GetCalls()["call1"].GetParameters()[0].GetSecret().GetPlainText() != "12345" {
        t.Error("Failed to hydrate WebsocketService secret")
    }

    // Test WebrtcService
    webrtcSvc := &configv1.UpstreamServiceConfig{}
	{
		svc := &configv1.WebrtcUpstreamService{}

		paramMapping := &configv1.WebrtcParameterMapping{}
		secretVal := &configv1.SecretValue{}
		secretVal.SetEnvironmentVariable("API_KEY")
		paramMapping.SetSecret(secretVal)

		callDef := &configv1.WebrtcCallDefinition{}
		callDef.SetParameters([]*configv1.WebrtcParameterMapping{paramMapping})

		svc.SetCalls(map[string]*configv1.WebrtcCallDefinition{
			"call1": callDef,
			"callNil": nil,
		})

		webrtcSvc.SetWebrtcService(svc)
	}
    HydrateSecretsInService(webrtcSvc, secrets)
    if webrtcSvc.GetWebrtcService().GetCalls()["call1"].GetParameters()[0].GetSecret().GetPlainText() != "12345" {
        t.Error("Failed to hydrate WebrtcService secret")
    }

    // Test HTTP Service with nil call
    httpSvc := &configv1.UpstreamServiceConfig{}
	{
		svc := &configv1.HttpUpstreamService{}
		svc.SetCalls(map[string]*configv1.HttpCallDefinition{
			"callNil": nil,
		})
		httpSvc.SetHttpService(svc)
	}
    HydrateSecretsInService(httpSvc, secrets)
}

func TestStripSecretsFromService_Filesystem_Vector_More(t *testing.T) {
     // Test Filesystem S3
    fsS3 := &configv1.UpstreamServiceConfig{}
	{
		s3 := &configv1.S3Fs{}
		s3.SetSecretAccessKey("secret")
		s3.SetSessionToken("token")

		svc := &configv1.FilesystemUpstreamService{}
		svc.SetS3(s3)
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
    fsSftp := &configv1.UpstreamServiceConfig{}
	{
		sftp := &configv1.SftpFs{}
		sftp.SetPassword("password")

		svc := &configv1.FilesystemUpstreamService{}
		svc.SetSftp(sftp)
		fsSftp.SetFilesystemService(svc)
	}
    StripSecretsFromService(fsSftp)
    if fsSftp.GetFilesystemService().GetSftp().GetPassword() != "" {
        t.Error("Failed to strip SFTP password")
    }

    // Test Vector Pinecone
    vecPine := &configv1.UpstreamServiceConfig{}
	{
		pinecone := &configv1.PineconeVectorDB{}
		pinecone.SetApiKey("secret")

		svc := &configv1.VectorUpstreamService{}
		svc.SetPinecone(pinecone)
		vecPine.SetVectorService(svc)
	}

    StripSecretsFromService(vecPine)
    if vecPine.GetVectorService().GetPinecone().GetApiKey() != "" {
        t.Error("Failed to strip Pinecone API key")
    }

    // Test Vector Milvus
    vecMilvus := &configv1.UpstreamServiceConfig{}
	{
		milvus := &configv1.MilvusVectorDB{}
		milvus.SetApiKey("secret")
		milvus.SetPassword("pass")

		svc := &configv1.VectorUpstreamService{}
		svc.SetMilvus(milvus)
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
