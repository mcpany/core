// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
			svc: &configv1.UpstreamServiceConfig{
				ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
					GrpcService: &configv1.GrpcUpstreamService{
						Address: proto.String("localhost:50051"),
					},
				},
			},
		},
		{
			name: "OpenapiService",
			svc: &configv1.UpstreamServiceConfig{
				ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
					OpenapiService: &configv1.OpenapiUpstreamService{
						SpecSource: &configv1.OpenapiUpstreamService_SpecUrl{
							SpecUrl: "http://example.com/spec.json",
						},
					},
				},
			},
		},
		{
			name: "McpService",
			svc: &configv1.UpstreamServiceConfig{
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						Calls: map[string]*configv1.MCPCallDefinition{
							"call1": {
								// Empty is fine
							},
						},
					},
				},
			},
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
    stdio := &configv1.UpstreamServiceConfig{
        ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
            McpService: &configv1.McpUpstreamService{
                ConnectionType: &configv1.McpUpstreamService_StdioConnection{
                    StdioConnection: &configv1.McpStdioConnection{
                        Env: map[string]*configv1.SecretValue{
                            "KEY": {
                                Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
                            },
                        },
                    },
                },
            },
        },
    }
    StripSecretsFromService(stdio)
    if stdio.GetMcpService().GetStdioConnection().GetEnv()["KEY"].GetPlainText() != "" {
        t.Error("Failed to strip McpService Stdio Env secret")
    }

    // Bundle
    bundle := &configv1.UpstreamServiceConfig{
        ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
            McpService: &configv1.McpUpstreamService{
                ConnectionType: &configv1.McpUpstreamService_BundleConnection{
                    BundleConnection: &configv1.McpBundleConnection{
                        Env: map[string]*configv1.SecretValue{
                            "KEY": {
                                Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
                            },
                        },
                    },
                },
            },
        },
    }
    StripSecretsFromService(bundle)
    if bundle.GetMcpService().GetBundleConnection().GetEnv()["KEY"].GetPlainText() != "" {
        t.Error("Failed to strip McpService Bundle Env secret")
    }
}

func TestStripSecretsFromProfile_Coverage(t *testing.T) {
	StripSecretsFromProfile(nil)

	profile := &configv1.ProfileDefinition{
		Secrets: map[string]*configv1.SecretValue{
			"key1": {
				Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
			},
		},
	}
	StripSecretsFromProfile(profile)
	if profile.Secrets["key1"].GetValue() != nil {
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
    mcpSvc := &configv1.UpstreamServiceConfig{
        ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
            McpService: &configv1.McpUpstreamService{
                ConnectionType: &configv1.McpUpstreamService_StdioConnection{
                    StdioConnection: &configv1.McpStdioConnection{
                        Env: map[string]*configv1.SecretValue{
                            "KEY": {
                                Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "API_KEY"},
                            },
                        },
                    },
                },
            },
        },
    }
    HydrateSecretsInService(mcpSvc, secrets)
    if mcpSvc.GetMcpService().GetStdioConnection().GetEnv()["KEY"].GetPlainText() != "12345" {
        t.Error("Failed to hydrate McpService Stdio Env")
    }

    // Test McpService Bundle
    mcpBundleSvc := &configv1.UpstreamServiceConfig{
        ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
            McpService: &configv1.McpUpstreamService{
                ConnectionType: &configv1.McpUpstreamService_BundleConnection{
                    BundleConnection: &configv1.McpBundleConnection{
                        Env: map[string]*configv1.SecretValue{
                            "KEY": {
                                Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "API_KEY"},
                            },
                        },
                    },
                },
            },
        },
    }
    HydrateSecretsInService(mcpBundleSvc, secrets)
    if mcpBundleSvc.GetMcpService().GetBundleConnection().GetEnv()["KEY"].GetPlainText() != "12345" {
        t.Error("Failed to hydrate McpService Bundle Env")
    }

    // Test CommandLineService
    cmdSvc := &configv1.UpstreamServiceConfig{
        ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
            CommandLineService: &configv1.CommandLineUpstreamService{
                Env: map[string]*configv1.SecretValue{
                     "KEY": {
                        Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "API_KEY"},
                    },
                },
                ContainerEnvironment: &configv1.ContainerEnvironment{
                    Env: map[string]*configv1.SecretValue{
                         "KEY_CONTAINER": {
                            Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "API_KEY"},
                        },
                    },
                },
            },
        },
    }
    HydrateSecretsInService(cmdSvc, secrets)
     if cmdSvc.GetCommandLineService().GetEnv()["KEY"].GetPlainText() != "12345" {
        t.Error("Failed to hydrate CommandLineService Env")
    }
     if cmdSvc.GetCommandLineService().GetContainerEnvironment().GetEnv()["KEY_CONTAINER"].GetPlainText() != "12345" {
        t.Error("Failed to hydrate CommandLineService Container Env")
    }

    // Test WebsocketService
    wsSvc := &configv1.UpstreamServiceConfig{
        ServiceConfig: &configv1.UpstreamServiceConfig_WebsocketService{
            WebsocketService: &configv1.WebsocketUpstreamService{
                Calls: map[string]*configv1.WebsocketCallDefinition{
                    "call1": {
                        Parameters: []*configv1.WebsocketParameterMapping{
                            {
                                Secret: &configv1.SecretValue{
                                     Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "API_KEY"},
                                },
                            },
                        },
                    },
                    "callNil": nil, // Test nil call
                },
            },
        },
    }
    HydrateSecretsInService(wsSvc, secrets)
    if wsSvc.GetWebsocketService().GetCalls()["call1"].GetParameters()[0].GetSecret().GetPlainText() != "12345" {
        t.Error("Failed to hydrate WebsocketService secret")
    }

    // Test WebrtcService
    webrtcSvc := &configv1.UpstreamServiceConfig{
        ServiceConfig: &configv1.UpstreamServiceConfig_WebrtcService{
            WebrtcService: &configv1.WebrtcUpstreamService{
                 Calls: map[string]*configv1.WebrtcCallDefinition{
                    "call1": {
                        Parameters: []*configv1.WebrtcParameterMapping{
                            {
                                Secret: &configv1.SecretValue{
                                     Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "API_KEY"},
                                },
                            },
                        },
                    },
                    "callNil": nil, // Test nil call
                },
            },
        },
    }
    HydrateSecretsInService(webrtcSvc, secrets)
    if webrtcSvc.GetWebrtcService().GetCalls()["call1"].GetParameters()[0].GetSecret().GetPlainText() != "12345" {
        t.Error("Failed to hydrate WebrtcService secret")
    }

    // Test HTTP Service with nil call
    httpSvc := &configv1.UpstreamServiceConfig{
        ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
            HttpService: &configv1.HttpUpstreamService{
                Calls: map[string]*configv1.HttpCallDefinition{
                    "callNil": nil,
                },
            },
        },
    }
    HydrateSecretsInService(httpSvc, secrets)
}

func TestStripSecretsFromService_Filesystem_Vector_More(t *testing.T) {
     // Test Filesystem S3
    fsS3 := &configv1.UpstreamServiceConfig{
        ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
            FilesystemService: &configv1.FilesystemUpstreamService{
                FilesystemType: &configv1.FilesystemUpstreamService_S3{
                    S3: &configv1.S3Fs{
                        SecretAccessKey: proto.String("secret"),
                        SessionToken: proto.String("token"),
                    },
                },
            },
        },
    }
    StripSecretsFromService(fsS3)
    if fsS3.GetFilesystemService().GetS3().GetSecretAccessKey() != "" {
        t.Error("Failed to strip S3 secret")
    }
     if fsS3.GetFilesystemService().GetS3().GetSessionToken() != "" {
        t.Error("Failed to strip S3 token")
    }

     // Test Filesystem SFTP
    fsSftp := &configv1.UpstreamServiceConfig{
        ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
            FilesystemService: &configv1.FilesystemUpstreamService{
                FilesystemType: &configv1.FilesystemUpstreamService_Sftp{
                    Sftp: &configv1.SftpFs{
                        Password: proto.String("password"),
                    },
                },
            },
        },
    }
    StripSecretsFromService(fsSftp)
    if fsSftp.GetFilesystemService().GetSftp().GetPassword() != "" {
        t.Error("Failed to strip SFTP password")
    }

    // Test Vector Pinecone
    vecPine := &configv1.UpstreamServiceConfig{
        ServiceConfig: &configv1.UpstreamServiceConfig_VectorService{
            VectorService: &configv1.VectorUpstreamService{
                VectorDbType: &configv1.VectorUpstreamService_Pinecone{
                    Pinecone: &configv1.PineconeVectorDB{
                        ApiKey: proto.String("secret"),
                    },
                },
            },
        },
    }

    StripSecretsFromService(vecPine)
    if vecPine.GetVectorService().GetPinecone().GetApiKey() != "" {
        t.Error("Failed to strip Pinecone API key")
    }

    // Test Vector Milvus
    vecMilvus := &configv1.UpstreamServiceConfig{
        ServiceConfig: &configv1.UpstreamServiceConfig_VectorService{
            VectorService: &configv1.VectorUpstreamService{
                VectorDbType: &configv1.VectorUpstreamService_Milvus{
                    Milvus: &configv1.MilvusVectorDB{
                        ApiKey: proto.String("secret"),
                        Password: proto.String("pass"),
                    },
                },
            },
        },
    }

    StripSecretsFromService(vecMilvus)
    if vecMilvus.GetVectorService().GetMilvus().GetApiKey() != "" {
        t.Error("Failed to strip Milvus API key")
    }
    if vecMilvus.GetVectorService().GetMilvus().GetPassword() != "" {
        t.Error("Failed to strip Milvus password")
    }
}
