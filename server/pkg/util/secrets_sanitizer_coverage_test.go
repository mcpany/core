// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func httpMethodPtr(v configv1.HttpCallDefinition_HttpMethod) *configv1.HttpCallDefinition_HttpMethod {
	return &v
}

func TestStripSecretsFromService_Coverage(t *testing.T) {
	// Test stripping secrets from various service types to ensure 100% coverage of switch cases

	// 1. GrpcService
	grpcSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
			GrpcService: &configv1.GrpcUpstreamService{
				Address: proto.String("localhost:50051"),
			},
		},
	}
	StripSecretsFromService(grpcSvc)
	assert.NotNil(t, grpcSvc)

	// 2. OpenapiService
	openapiSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
			OpenapiService: &configv1.OpenapiUpstreamService{
				SpecSource: &configv1.OpenapiUpstreamService_SpecUrl{
					SpecUrl: "http://localhost/openapi.json",
				},
			},
		},
	}
	StripSecretsFromService(openapiSvc)
	assert.NotNil(t, openapiSvc)

	// 3. McpService with Calls
	mcpSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				Calls: map[string]*configv1.MCPCallDefinition{
					"call1": {
						Id: proto.String("call1"),
					},
				},
			},
		},
	}
	StripSecretsFromService(mcpSvc)
	assert.NotNil(t, mcpSvc)

	// 4. CommandLineService
	cmdSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
			CommandLineService: &configv1.CommandLineUpstreamService{
				Command: proto.String("echo"),
				Env: map[string]*configv1.SecretValue{
					"FOO": {Value: &configv1.SecretValue_PlainText{PlainText: "secret"}},
				},
				Calls: map[string]*configv1.CommandLineCallDefinition{
					"call_nil": nil,
				},
			},
		},
	}
	StripSecretsFromService(cmdSvc)
	// Verify environment variable secret was stripped
	envVar := cmdSvc.GetCommandLineService().Env["FOO"]
	assert.Nil(t, envVar.Value, "Plain text secret in Env should be stripped")

	// 5. HttpService with Calls
	httpSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Calls: map[string]*configv1.HttpCallDefinition{
					"call1": {
						Id:     proto.String("call1"),
						Method: httpMethodPtr(configv1.HttpCallDefinition_HTTP_METHOD_GET),
						Parameters: []*configv1.HttpParameterMapping{
							{
								Secret: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "secret"}},
							},
						},
					},
					"call_nil": nil,
				},
			},
		},
	}
	StripSecretsFromService(httpSvc)
	// Verify secret stripped
	secret := httpSvc.GetHttpService().Calls["call1"].Parameters[0].Secret
	assert.Nil(t, secret.Value, "Plain text secret in HttpCall should be stripped")

	// 6. WebsocketService
	wsSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_WebsocketService{
			WebsocketService: &configv1.WebsocketUpstreamService{
				Calls: map[string]*configv1.WebsocketCallDefinition{
					"call1": {
						Parameters: []*configv1.WebsocketParameterMapping{
							{
								Secret: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "secret"}},
							},
						},
					},
					"call_nil": nil,
				},
			},
		},
	}
	StripSecretsFromService(wsSvc)
	assert.Nil(t, wsSvc.GetWebsocketService().Calls["call1"].Parameters[0].Secret.Value)

	// 7. WebrtcService
	webrtcSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_WebrtcService{
			WebrtcService: &configv1.WebrtcUpstreamService{
				Calls: map[string]*configv1.WebrtcCallDefinition{
					"call1": {
						Parameters: []*configv1.WebrtcParameterMapping{
							{
								Secret: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "secret"}},
							},
						},
					},
					"call_nil": nil,
				},
			},
		},
	}
	StripSecretsFromService(webrtcSvc)
	assert.Nil(t, webrtcSvc.GetWebrtcService().Calls["call1"].Parameters[0].Secret.Value)

	// 8. Nil checks
	StripSecretsFromService(nil)
	StripSecretsFromAuth(nil)
	StripSecretsFromProfile(nil)
	StripSecretsFromCollection(nil)
}

func TestStripSecretsFromFilesystem_Coverage(t *testing.T) {
	// S3
	s3Svc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_S3{
					S3: &configv1.S3Fs{
						SecretAccessKey: proto.String("secret"),
						SessionToken:    proto.String("token"),
					},
				},
			},
		},
	}
	StripSecretsFromService(s3Svc)
	assert.Equal(t, "", *s3Svc.GetFilesystemService().GetS3().SecretAccessKey)
	assert.Equal(t, "", *s3Svc.GetFilesystemService().GetS3().SessionToken)

	// SFTP
	sftpSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_Sftp{
					Sftp: &configv1.SftpFs{
						Password: proto.String("secret"),
					},
				},
			},
		},
	}
	StripSecretsFromService(sftpSvc)
	assert.Equal(t, "", *sftpSvc.GetFilesystemService().GetSftp().Password)
}

func TestStripSecretsFromVector_Coverage(t *testing.T) {
	// Pinecone
	pinecone := &configv1.UpstreamServiceConfig{
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
	StripSecretsFromService(pinecone)
	assert.Equal(t, "", *pinecone.GetVectorService().GetPinecone().ApiKey)

	// Milvus
	milvus := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_VectorService{
			VectorService: &configv1.VectorUpstreamService{
				VectorDbType: &configv1.VectorUpstreamService_Milvus{
					Milvus: &configv1.MilvusVectorDB{
						ApiKey:   proto.String("secret"),
						Password: proto.String("pass"),
					},
				},
			},
		},
	}
	StripSecretsFromService(milvus)
	assert.Equal(t, "", *milvus.GetVectorService().GetMilvus().ApiKey)
	assert.Equal(t, "", *milvus.GetVectorService().GetMilvus().Password)
}

func TestHydrateSecrets_Coverage(t *testing.T) {
	secrets := map[string]*configv1.SecretValue{
		"MY_SECRET": {Value: &configv1.SecretValue_PlainText{PlainText: "resolved"}},
	}

	// 1. Env Var hydration
	svc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
			CommandLineService: &configv1.CommandLineUpstreamService{
				Env: map[string]*configv1.SecretValue{
					"TARGET":  {Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MY_SECRET"}},
					"MISSING": {Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "NOT_FOUND"}},
					"PLAIN":   {Value: &configv1.SecretValue_PlainText{PlainText: "plain"}},
				},
			},
		},
	}
	HydrateSecretsInService(svc, secrets)

	target := svc.GetCommandLineService().Env["TARGET"]
	assert.Equal(t, "resolved", target.GetPlainText())

	missing := svc.GetCommandLineService().Env["MISSING"]
	assert.Equal(t, "NOT_FOUND", missing.GetEnvironmentVariable()) // Should remain as env var

	plain := svc.GetCommandLineService().Env["PLAIN"]
	assert.Equal(t, "plain", plain.GetPlainText())

	// 2. HTTP Service hydration
	httpSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Calls: map[string]*configv1.HttpCallDefinition{
					"c1": {
						Parameters: []*configv1.HttpParameterMapping{
							{
								Secret: &configv1.SecretValue{Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MY_SECRET"}},
							},
							{
								// Nil secret
								Secret: nil,
							},
						},
					},
				},
			},
		},
	}
	HydrateSecretsInService(httpSvc, secrets)
	resolved := httpSvc.GetHttpService().Calls["c1"].Parameters[0].Secret
	assert.Equal(t, "resolved", resolved.GetPlainText())

	// 3. Websocket Service hydration
	wsSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_WebsocketService{
			WebsocketService: &configv1.WebsocketUpstreamService{
				Calls: map[string]*configv1.WebsocketCallDefinition{
					"c1": {
						Parameters: []*configv1.WebsocketParameterMapping{
							{
								Secret: &configv1.SecretValue{Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MY_SECRET"}},
							},
						},
					},
				},
			},
		},
	}
	HydrateSecretsInService(wsSvc, secrets)
	wsResolved := wsSvc.GetWebsocketService().Calls["c1"].Parameters[0].Secret
	assert.Equal(t, "resolved", wsResolved.GetPlainText())

	// 4. Nil checks
	HydrateSecretsInService(nil, secrets)
	HydrateSecretsInService(svc, nil)
}

func TestStripSecretsFromAuth_Coverage(t *testing.T) {
	// 1. API Key with Verification Value
	auth := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: &configv1.APIKeyAuth{
				Value:             &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "secret"}},
				VerificationValue: proto.String("verify"),
			},
		},
	}
	StripSecretsFromAuth(auth)
	assert.Nil(t, auth.GetApiKey().Value.Value)
	assert.Equal(t, "", *auth.GetApiKey().VerificationValue)

	// 2. Bearer Token
	bearer := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_BearerToken{
			BearerToken: &configv1.BearerTokenAuth{
				Token: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "secret"}},
			},
		},
	}
	StripSecretsFromAuth(bearer)
	assert.Nil(t, bearer.GetBearerToken().Token.Value)

	// 3. Basic Auth
	basic := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_BasicAuth{
			BasicAuth: &configv1.BasicAuth{
				Password: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "secret"}},
			},
		},
	}
	StripSecretsFromAuth(basic)
	assert.Nil(t, basic.GetBasicAuth().Password.Value)

	// 4. OAuth2
	oauth := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{
				ClientId:     &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "id"}},
				ClientSecret: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "secret"}},
			},
		},
	}
	StripSecretsFromAuth(oauth)
	assert.Nil(t, oauth.GetOauth2().ClientId.Value)
	assert.Nil(t, oauth.GetOauth2().ClientSecret.Value)
}
