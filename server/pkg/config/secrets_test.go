// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestStripSecretsFromService(t *testing.T) {
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		UpstreamAuth: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					ParamName: proto.String("X-API-Key"),
					Value: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{PlainText: "secret-key"},
					},
				},
			},
		},
	}

	StripSecretsFromService(svc)

	assert.NotNil(t, svc.UpstreamAuth)
	assert.NotNil(t, svc.UpstreamAuth.GetApiKey())
	assert.NotNil(t, svc.UpstreamAuth.GetApiKey().Value)
	assert.Nil(t, svc.UpstreamAuth.GetApiKey().Value.Value, "Plain text secret should be cleared")
}

func TestStripSecretsFromProfile(t *testing.T) {
	profile := &configv1.ProfileDefinition{
		Name: proto.String("test-profile"),
		Secrets: map[string]*configv1.SecretValue{
			"TEST_SECRET": {Value: &configv1.SecretValue_PlainText{PlainText: "secret-value"}},
		},
	}

	StripSecretsFromProfile(profile)

	secret := profile.Secrets["TEST_SECRET"]
	assert.NotNil(t, secret)
	assert.Nil(t, secret.Value, "Plain text secret should be cleared")
}

func TestStripSecretsFromCollection(t *testing.T) {
	collection := &configv1.Collection{
		Name: proto.String("test-collection"),
		Services: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("svc1"),
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_BasicAuth{
						BasicAuth: &configv1.BasicAuth{
							Username: proto.String("user"),
							Password: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{PlainText: "secret-password"},
							},
						},
					},
				},
			},
		},
	}

	StripSecretsFromCollection(collection)

	svc := collection.Services[0]
	assert.NotNil(t, svc.UpstreamAuth)
	assert.Nil(t, svc.UpstreamAuth.GetBasicAuth().Password.Value, "Plain text secret should be cleared")
}

func TestHydrateSecretsInService(t *testing.T) {
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		UpstreamAuth: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					ParamName: proto.String("X-API-Key"),
					Value: &configv1.SecretValue{
						Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "API_KEY_VAR"},
					},
				},
			},
		},
	}

	secrets := map[string]*configv1.SecretValue{
		"API_KEY_VAR": {Value: &configv1.SecretValue_PlainText{PlainText: "resolved-secret"}},
	}

	HydrateSecretsInService(svc, secrets)

	val := svc.UpstreamAuth.GetApiKey().Value.Value.(*configv1.SecretValue_PlainText)
	assert.Equal(t, "resolved-secret", val.PlainText)
}

func TestHydrateSecretsInService_HttpService(t *testing.T) {
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-http-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Calls: map[string]*configv1.HttpCallDefinition{
					"call1": {
						Parameters: []*configv1.HttpParameterMapping{
							{
								Schema: &configv1.ParameterSchema{Name: proto.String("apiKey")},
								Secret: &configv1.SecretValue{
									Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "API_KEY_VAR"},
								},
							},
						},
					},
				},
			},
		},
	}

	secrets := map[string]*configv1.SecretValue{
		"API_KEY_VAR": {Value: &configv1.SecretValue_PlainText{PlainText: "resolved-secret"}},
	}

	HydrateSecretsInService(svc, secrets)

	httpSvc := svc.GetHttpService()
	assert.NotNil(t, httpSvc)
	call := httpSvc.Calls["call1"]
	assert.NotNil(t, call)
	param := call.Parameters[0]
	assert.NotNil(t, param.Secret)

	val, ok := param.Secret.Value.(*configv1.SecretValue_PlainText)
	if assert.True(t, ok, "Secret value should be PlainText after hydration") {
		assert.Equal(t, "resolved-secret", val.PlainText)
	}
}

func TestHydrateSecretsInService_WebsocketService(t *testing.T) {
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-ws-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_WebsocketService{
			WebsocketService: &configv1.WebsocketUpstreamService{
				Calls: map[string]*configv1.WebsocketCallDefinition{
					"call1": {
						Parameters: []*configv1.WebsocketParameterMapping{
							{
								Schema: &configv1.ParameterSchema{Name: proto.String("token")},
								Secret: &configv1.SecretValue{
									Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "WS_TOKEN"},
								},
							},
						},
					},
				},
			},
		},
	}

	secrets := map[string]*configv1.SecretValue{
		"WS_TOKEN": {Value: &configv1.SecretValue_PlainText{PlainText: "resolved-token"}},
	}

	HydrateSecretsInService(svc, secrets)

	wsSvc := svc.GetWebsocketService()
	assert.NotNil(t, wsSvc)
	call := wsSvc.Calls["call1"]
	assert.NotNil(t, call)
	param := call.Parameters[0]
	assert.NotNil(t, param.Secret)

	val, ok := param.Secret.Value.(*configv1.SecretValue_PlainText)
	if assert.True(t, ok, "Secret value should be PlainText after hydration") {
		assert.Equal(t, "resolved-token", val.PlainText)
	}
}

func TestHydrateSecretsInService_WebrtcService(t *testing.T) {
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-webrtc-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_WebrtcService{
			WebrtcService: &configv1.WebrtcUpstreamService{
				Calls: map[string]*configv1.WebrtcCallDefinition{
					"call1": {
						Parameters: []*configv1.WebrtcParameterMapping{
							{
								Schema: &configv1.ParameterSchema{Name: proto.String("secret")},
								Secret: &configv1.SecretValue{
									Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "RTC_SECRET"},
								},
							},
						},
					},
				},
			},
		},
	}

	secrets := map[string]*configv1.SecretValue{
		"RTC_SECRET": {Value: &configv1.SecretValue_PlainText{PlainText: "resolved-secret"}},
	}

	HydrateSecretsInService(svc, secrets)

	rtcSvc := svc.GetWebrtcService()
	assert.NotNil(t, rtcSvc)
	call := rtcSvc.Calls["call1"]
	assert.NotNil(t, call)
	param := call.Parameters[0]
	assert.NotNil(t, param.Secret)

	val, ok := param.Secret.Value.(*configv1.SecretValue_PlainText)
	if assert.True(t, ok, "Secret value should be PlainText after hydration") {
		assert.Equal(t, "resolved-secret", val.PlainText)
	}
}

func TestHydrateSecretsInService_CommandLineService(t *testing.T) {
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-cmd-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
			CommandLineService: &configv1.CommandLineUpstreamService{
				Env: map[string]*configv1.SecretValue{
					"ENV_VAR": {
						Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "SECRET_ENV"},
					},
				},
				ContainerEnvironment: &configv1.ContainerEnvironment{
					Env: map[string]*configv1.SecretValue{
						"CONTAINER_VAR": {
							Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "SECRET_CONTAINER"},
						},
					},
				},
			},
		},
	}

	secrets := map[string]*configv1.SecretValue{
		"SECRET_ENV":       {Value: &configv1.SecretValue_PlainText{PlainText: "env-value"}},
		"SECRET_CONTAINER": {Value: &configv1.SecretValue_PlainText{PlainText: "container-value"}},
	}

	HydrateSecretsInService(svc, secrets)

	cmdSvc := svc.GetCommandLineService()
	assert.NotNil(t, cmdSvc)

	envVal, ok := cmdSvc.Env["ENV_VAR"].Value.(*configv1.SecretValue_PlainText)
	if assert.True(t, ok) {
		assert.Equal(t, "env-value", envVal.PlainText)
	}

	containerVal, ok := cmdSvc.ContainerEnvironment.Env["CONTAINER_VAR"].Value.(*configv1.SecretValue_PlainText)
	if assert.True(t, ok) {
		assert.Equal(t, "container-value", containerVal.PlainText)
	}
}

func TestHydrateSecretsInService_McpService(t *testing.T) {
	// Stdio
	svcStdio := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-mcp-stdio"),
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_StdioConnection{
					StdioConnection: &configv1.McpStdioConnection{
						Env: map[string]*configv1.SecretValue{
							"STDIO_VAR": {
								Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "SECRET_STDIO"},
							},
						},
					},
				},
			},
		},
	}

	// Bundle
	svcBundle := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-mcp-bundle"),
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_BundleConnection{
					BundleConnection: &configv1.McpBundleConnection{
						Env: map[string]*configv1.SecretValue{
							"BUNDLE_VAR": {
								Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "SECRET_BUNDLE"},
							},
						},
					},
				},
			},
		},
	}

	secrets := map[string]*configv1.SecretValue{
		"SECRET_STDIO":  {Value: &configv1.SecretValue_PlainText{PlainText: "stdio-value"}},
		"SECRET_BUNDLE": {Value: &configv1.SecretValue_PlainText{PlainText: "bundle-value"}},
	}

	HydrateSecretsInService(svcStdio, secrets)
	HydrateSecretsInService(svcBundle, secrets)

	stdioEnvVal, ok := svcStdio.GetMcpService().GetStdioConnection().Env["STDIO_VAR"].Value.(*configv1.SecretValue_PlainText)
	if assert.True(t, ok) {
		assert.Equal(t, "stdio-value", stdioEnvVal.PlainText)
	}

	bundleEnvVal, ok := svcBundle.GetMcpService().GetBundleConnection().Env["BUNDLE_VAR"].Value.(*configv1.SecretValue_PlainText)
	if assert.True(t, ok) {
		assert.Equal(t, "bundle-value", bundleEnvVal.PlainText)
	}
}

func TestStripSecretsFromAllServices(t *testing.T) {
	// Setup services with secrets
	services := []*configv1.UpstreamServiceConfig{
		{
			Name: proto.String("http"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Calls: map[string]*configv1.HttpCallDefinition{
						"c": {Parameters: []*configv1.HttpParameterMapping{{Secret: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "s"}}}}},
					},
				},
			},
		},
		{
			Name: proto.String("ws"),
			ServiceConfig: &configv1.UpstreamServiceConfig_WebsocketService{
				WebsocketService: &configv1.WebsocketUpstreamService{
					Calls: map[string]*configv1.WebsocketCallDefinition{
						"c": {Parameters: []*configv1.WebsocketParameterMapping{{Secret: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "s"}}}}},
					},
				},
			},
		},
		{
			Name: proto.String("webrtc"),
			ServiceConfig: &configv1.UpstreamServiceConfig_WebrtcService{
				WebrtcService: &configv1.WebrtcUpstreamService{
					Calls: map[string]*configv1.WebrtcCallDefinition{
						"c": {Parameters: []*configv1.WebrtcParameterMapping{{Secret: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "s"}}}}},
					},
				},
			},
		},
		{
			Name: proto.String("mcp"),
			ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
				McpService: &configv1.McpUpstreamService{
					ConnectionType: &configv1.McpUpstreamService_StdioConnection{
						StdioConnection: &configv1.McpStdioConnection{
							Env: map[string]*configv1.SecretValue{"k": {Value: &configv1.SecretValue_PlainText{PlainText: "s"}}},
						},
					},
				},
			},
		},
		{
			Name: proto.String("vector"),
			ServiceConfig: &configv1.UpstreamServiceConfig_VectorService{
				VectorService: &configv1.VectorUpstreamService{
					VectorDbType: &configv1.VectorUpstreamService_Pinecone{
						Pinecone: &configv1.PineconeVectorDB{ApiKey: proto.String("s")},
					},
				},
			},
		},
		{
			Name: proto.String("fs"),
			ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
				FilesystemService: &configv1.FilesystemUpstreamService{
					FilesystemType: &configv1.FilesystemUpstreamService_S3{
						S3: &configv1.S3Fs{SecretAccessKey: proto.String("s")},
					},
				},
			},
		},
	}

	for _, svc := range services {
		StripSecretsFromService(svc)
	}

	// Verify secrets are gone
	assert.Nil(t, services[0].GetHttpService().Calls["c"].Parameters[0].Secret.Value)
	assert.Nil(t, services[1].GetWebsocketService().Calls["c"].Parameters[0].Secret.Value)
	assert.Nil(t, services[2].GetWebrtcService().Calls["c"].Parameters[0].Secret.Value)
	assert.Nil(t, services[3].GetMcpService().GetStdioConnection().Env["k"].Value)
	assert.Equal(t, "", services[4].GetVectorService().GetPinecone().GetApiKey())
	assert.Equal(t, "", services[5].GetFilesystemService().GetS3().GetSecretAccessKey())
}

func TestStripSecretsFromGrpcAndOpenapi(t *testing.T) {
	// These currently don't do much, but we test for coverage
	svcGrpc := &configv1.UpstreamServiceConfig{
		Name: proto.String("grpc"),
		ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
			GrpcService: &configv1.GrpcUpstreamService{},
		},
	}
	StripSecretsFromService(svcGrpc)

	svcOpenapi := &configv1.UpstreamServiceConfig{
		Name: proto.String("openapi"),
		ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
			OpenapiService: &configv1.OpenapiUpstreamService{},
		},
	}
	StripSecretsFromService(svcOpenapi)
}

func TestStripSecretsFromHookAndCache(t *testing.T) {
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("hook-cache"),
		PreCallHooks: []*configv1.CallHook{
			{
				HookConfig: &configv1.CallHook_Webhook{
					Webhook: &configv1.WebhookConfig{WebhookSecret: "secret"},
				},
			},
		},
		Cache: &configv1.CacheConfig{
			SemanticConfig: &configv1.SemanticCacheConfig{
				ApiKey: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "s"}},
				ProviderConfig: &configv1.SemanticCacheConfig_Openai{
					Openai: &configv1.OpenAIEmbeddingProviderConfig{
						ApiKey: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "s"}},
					},
				},
			},
		},
	}

	StripSecretsFromService(svc)

	assert.Equal(t, "", svc.PreCallHooks[0].GetWebhook().WebhookSecret)
	assert.Nil(t, svc.Cache.SemanticConfig.ApiKey.Value)
	assert.Nil(t, svc.Cache.SemanticConfig.GetOpenai().ApiKey.Value)
}

func TestHydrateSecretsInAuth_Extended(t *testing.T) {
	auth := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_BearerToken{
			BearerToken: &configv1.BearerTokenAuth{
				Token: &configv1.SecretValue{
					Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "TOKEN_VAR"},
				},
			},
		},
	}

	secrets := map[string]*configv1.SecretValue{
		"TOKEN_VAR": {Value: &configv1.SecretValue_PlainText{PlainText: "token"}},
	}

	hydrateSecretsInAuth(auth, secrets)
	assert.Equal(t, "token", auth.GetBearerToken().Token.Value.(*configv1.SecretValue_PlainText).PlainText)

	authBasic := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_BasicAuth{
			BasicAuth: &configv1.BasicAuth{
				Password: &configv1.SecretValue{
					Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "PASS_VAR"},
				},
			},
		},
	}
	secrets["PASS_VAR"] = &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "pass"}}
	hydrateSecretsInAuth(authBasic, secrets)
	assert.Equal(t, "pass", authBasic.GetBasicAuth().Password.Value.(*configv1.SecretValue_PlainText).PlainText)

	authOauth := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{
				ClientId: &configv1.SecretValue{
					Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "ID_VAR"},
				},
				ClientSecret: &configv1.SecretValue{
					Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "SECRET_VAR"},
				},
			},
		},
	}
	secrets["ID_VAR"] = &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "id"}}
	secrets["SECRET_VAR"] = &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "secret"}}
	hydrateSecretsInAuth(authOauth, secrets)
	assert.Equal(t, "id", authOauth.GetOauth2().ClientId.Value.(*configv1.SecretValue_PlainText).PlainText)
	assert.Equal(t, "secret", authOauth.GetOauth2().ClientSecret.Value.(*configv1.SecretValue_PlainText).PlainText)
}

func TestCoverageForEmptyStripFunctions(t *testing.T) {
	// Call empty functions to ensure coverage
	stripSecretsFromGrpcService(nil)
	stripSecretsFromOpenapiService(nil)
	stripSecretsFromMcpCall(nil)
}
