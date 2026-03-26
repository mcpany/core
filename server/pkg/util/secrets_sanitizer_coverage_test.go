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
	grpcSvc := configv1.UpstreamServiceConfig_builder{
		GrpcService: configv1.GrpcUpstreamService_builder{
			Address: proto.String("localhost:50051"),
		}.Build(),
	}.Build()
	StripSecretsFromService(grpcSvc)
	assert.NotNil(t, grpcSvc)

	// 2. OpenapiService
	openapiSvc := configv1.UpstreamServiceConfig_builder{
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			SpecUrl: proto.String("http://localhost/openapi.json"),
		}.Build(),
	}.Build()
	StripSecretsFromService(openapiSvc)
	assert.NotNil(t, openapiSvc)

	// 3. McpService with Calls
	mcpSvc := configv1.UpstreamServiceConfig_builder{
		McpService: configv1.McpUpstreamService_builder{
			Calls: map[string]*configv1.MCPCallDefinition{
				"call1": configv1.MCPCallDefinition_builder{
					Id: proto.String("call1"),
				}.Build(),
			},
		}.Build(),
	}.Build()
	StripSecretsFromService(mcpSvc)
	assert.NotNil(t, mcpSvc)

	// 4. CommandLineService
	cmdSvc := configv1.UpstreamServiceConfig_builder{
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: proto.String("echo"),
			Env: map[string]*configv1.SecretValue{
				"FOO": configv1.SecretValue_builder{PlainText: proto.String("secret")}.Build(),
			},
			Calls: map[string]*configv1.CommandLineCallDefinition{
				"call_nil": nil,
			},
		}.Build(),
	}.Build()
	StripSecretsFromService(cmdSvc)
	// Verify environment variable secret was stripped
	envVar := cmdSvc.GetCommandLineService().GetEnv()["FOO"]
	assert.False(t, envVar.HasValue(), "Plain text secret in Env should be stripped")

	// 5. HttpService with Calls
	httpSvc := configv1.UpstreamServiceConfig_builder{
		HttpService: configv1.HttpUpstreamService_builder{
			Calls: map[string]*configv1.HttpCallDefinition{
				"call1": configv1.HttpCallDefinition_builder{
					Id:     proto.String("call1"),
					Method: configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
					Parameters: []*configv1.HttpParameterMapping{
						configv1.HttpParameterMapping_builder{
							Secret: configv1.SecretValue_builder{PlainText: proto.String("secret")}.Build(),
						}.Build(),
					},
				}.Build(),
				"call_nil": nil,
			},
		}.Build(),
	}.Build()
	StripSecretsFromService(httpSvc)
	// Verify secret stripped
	secret := httpSvc.GetHttpService().GetCalls()["call1"].GetParameters()[0].GetSecret()
	assert.False(t, secret.HasValue(), "Plain text secret in HttpCall should be stripped")

	// 6. WebsocketService
	wsSvc := configv1.UpstreamServiceConfig_builder{
		WebsocketService: configv1.WebsocketUpstreamService_builder{
			Calls: map[string]*configv1.WebsocketCallDefinition{
				"call1": configv1.WebsocketCallDefinition_builder{
					Parameters: []*configv1.WebsocketParameterMapping{
						configv1.WebsocketParameterMapping_builder{
							Secret: configv1.SecretValue_builder{PlainText: proto.String("secret")}.Build(),
						}.Build(),
					},
				}.Build(),
				"call_nil": nil,
			},
		}.Build(),
	}.Build()
	StripSecretsFromService(wsSvc)
	assert.False(t, wsSvc.GetWebsocketService().GetCalls()["call1"].GetParameters()[0].GetSecret().HasValue())

	// 7. WebrtcService
	webrtcSvc := configv1.UpstreamServiceConfig_builder{
		WebrtcService: configv1.WebrtcUpstreamService_builder{
			Calls: map[string]*configv1.WebrtcCallDefinition{
				"call1": configv1.WebrtcCallDefinition_builder{
					Parameters: []*configv1.WebrtcParameterMapping{
						configv1.WebrtcParameterMapping_builder{
							Secret: configv1.SecretValue_builder{PlainText: proto.String("secret")}.Build(),
						}.Build(),
					},
				}.Build(),
				"call_nil": nil,
			},
		}.Build(),
	}.Build()
	StripSecretsFromService(webrtcSvc)
	assert.False(t, webrtcSvc.GetWebrtcService().GetCalls()["call1"].GetParameters()[0].GetSecret().HasValue())

	// 8. Nil checks
	StripSecretsFromService(nil)
	StripSecretsFromAuth(nil)
	StripSecretsFromProfile(nil)
	StripSecretsFromCollection(nil)
}

func TestStripSecretsFromFilesystem_Coverage(t *testing.T) {
	// S3
	s3Svc := configv1.UpstreamServiceConfig_builder{
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			S3: configv1.S3Fs_builder{
				SecretAccessKey: proto.String("secret"),
				SessionToken:    proto.String("token"),
			}.Build(),
		}.Build(),
	}.Build()
	StripSecretsFromService(s3Svc)
	assert.Equal(t, "", s3Svc.GetFilesystemService().GetS3().GetSecretAccessKey())
	assert.Equal(t, "", s3Svc.GetFilesystemService().GetS3().GetSessionToken())

	// SFTP
	sftpSvc := configv1.UpstreamServiceConfig_builder{
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			Sftp: configv1.SftpFs_builder{
				Password: proto.String("secret"),
			}.Build(),
		}.Build(),
	}.Build()
	StripSecretsFromService(sftpSvc)
	assert.Equal(t, "", sftpSvc.GetFilesystemService().GetSftp().GetPassword())
}

func TestStripSecretsFromVector_Coverage(t *testing.T) {
	// Pinecone
	pinecone := configv1.UpstreamServiceConfig_builder{
		VectorService: configv1.VectorUpstreamService_builder{
			Pinecone: configv1.PineconeVectorDB_builder{
				ApiKey: proto.String("secret"),
			}.Build(),
		}.Build(),
	}.Build()
	StripSecretsFromService(pinecone)
	assert.Equal(t, "", pinecone.GetVectorService().GetPinecone().GetApiKey())

	// Milvus
	milvus := configv1.UpstreamServiceConfig_builder{
		VectorService: configv1.VectorUpstreamService_builder{
			Milvus: configv1.MilvusVectorDB_builder{
				ApiKey:   proto.String("secret"),
				Password: proto.String("pass"),
			}.Build(),
		}.Build(),
	}.Build()
	StripSecretsFromService(milvus)
	assert.Equal(t, "", milvus.GetVectorService().GetMilvus().GetApiKey())
	assert.Equal(t, "", milvus.GetVectorService().GetMilvus().GetPassword())
}

func TestHydrateSecrets_Coverage(t *testing.T) {
	secrets := map[string]*configv1.SecretValue{
		"MY_SECRET": configv1.SecretValue_builder{PlainText: proto.String("resolved")}.Build(),
	}

	// 1. Env Var hydration
	svc := configv1.UpstreamServiceConfig_builder{
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Env: map[string]*configv1.SecretValue{
				"TARGET":  configv1.SecretValue_builder{EnvironmentVariable: proto.String("MY_SECRET")}.Build(),
				"MISSING": configv1.SecretValue_builder{EnvironmentVariable: proto.String("NOT_FOUND")}.Build(),
				"PLAIN":   configv1.SecretValue_builder{PlainText: proto.String("plain")}.Build(),
			},
		}.Build(),
	}.Build()
	HydrateSecretsInService(svc, secrets)

	target := svc.GetCommandLineService().GetEnv()["TARGET"]
	assert.Equal(t, "resolved", target.GetPlainText())

	missing := svc.GetCommandLineService().GetEnv()["MISSING"]
	assert.Equal(t, "NOT_FOUND", missing.GetEnvironmentVariable()) // Should remain as env var

	plain := svc.GetCommandLineService().GetEnv()["PLAIN"]
	assert.Equal(t, "plain", plain.GetPlainText())

	// 2. HTTP Service hydration
	httpSvc := configv1.UpstreamServiceConfig_builder{
		HttpService: configv1.HttpUpstreamService_builder{
			Calls: map[string]*configv1.HttpCallDefinition{
				"c1": configv1.HttpCallDefinition_builder{
					Parameters: []*configv1.HttpParameterMapping{
						configv1.HttpParameterMapping_builder{
							Secret: configv1.SecretValue_builder{EnvironmentVariable: proto.String("MY_SECRET")}.Build(),
						}.Build(),
						configv1.HttpParameterMapping_builder{
							// Nil secret
							Secret: nil,
						}.Build(),
					},
				}.Build(),
			},
		}.Build(),
	}.Build()
	HydrateSecretsInService(httpSvc, secrets)
	resolved := httpSvc.GetHttpService().GetCalls()["c1"].GetParameters()[0].GetSecret()
	assert.Equal(t, "resolved", resolved.GetPlainText())

	// 3. Websocket Service hydration
	wsSvc := configv1.UpstreamServiceConfig_builder{
		WebsocketService: configv1.WebsocketUpstreamService_builder{
			Calls: map[string]*configv1.WebsocketCallDefinition{
				"c1": configv1.WebsocketCallDefinition_builder{
					Parameters: []*configv1.WebsocketParameterMapping{
						configv1.WebsocketParameterMapping_builder{
							Secret: configv1.SecretValue_builder{EnvironmentVariable: proto.String("MY_SECRET")}.Build(),
						}.Build(),
					},
				}.Build(),
			},
		}.Build(),
	}.Build()
	HydrateSecretsInService(wsSvc, secrets)
	wsResolved := wsSvc.GetWebsocketService().GetCalls()["c1"].GetParameters()[0].GetSecret()
	assert.Equal(t, "resolved", wsResolved.GetPlainText())

	// 4. Nil checks
	HydrateSecretsInService(nil, secrets)
	HydrateSecretsInService(svc, nil)
}

func TestStripSecretsFromAuth_Coverage(t *testing.T) {
	// 1. API Key with Verification Value
	auth := configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			Value:             configv1.SecretValue_builder{PlainText: proto.String("secret")}.Build(),
			VerificationValue: proto.String("verify"),
		}.Build(),
	}.Build()
	StripSecretsFromAuth(auth)
	assert.False(t, auth.GetApiKey().GetValue().HasValue())
	assert.Equal(t, "", auth.GetApiKey().GetVerificationValue())

	// 2. Bearer Token
	bearer := configv1.Authentication_builder{
		BearerToken: configv1.BearerTokenAuth_builder{
			Token: configv1.SecretValue_builder{PlainText: proto.String("secret")}.Build(),
		}.Build(),
	}.Build()
	StripSecretsFromAuth(bearer)
	assert.False(t, bearer.GetBearerToken().GetToken().HasValue())

	// 3. Basic Auth
	basic := configv1.Authentication_builder{
		BasicAuth: configv1.BasicAuth_builder{
			Password: configv1.SecretValue_builder{PlainText: proto.String("secret")}.Build(),
		}.Build(),
	}.Build()
	StripSecretsFromAuth(basic)
	assert.False(t, basic.GetBasicAuth().GetPassword().HasValue())

	// 4. OAuth2
	oauth := configv1.Authentication_builder{
		Oauth2: configv1.OAuth2Auth_builder{
			ClientId:     configv1.SecretValue_builder{PlainText: proto.String("id")}.Build(),
			ClientSecret: configv1.SecretValue_builder{PlainText: proto.String("secret")}.Build(),
		}.Build(),
	}.Build()
	StripSecretsFromAuth(oauth)
	assert.False(t, oauth.GetOauth2().GetClientId().HasValue())
	assert.False(t, oauth.GetOauth2().GetClientSecret().HasValue())
}
