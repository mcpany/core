// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestStripSecretsFromGrpcService(t *testing.T) {
	service := configv1.GrpcUpstreamService_builder{
		Address: proto.String("localhost:50051"),
		Calls: map[string]*configv1.GrpcCallDefinition{
			"call1": configv1.GrpcCallDefinition_builder{
				InputSchema:  nil,
				OutputSchema: nil,
			}.Build(),
		},
	}.Build()
	stripSecretsFromGrpcService(service)
	assert.Equal(t, "localhost:50051", service.GetAddress())
}

func TestStripSecretsFromOpenapiService(t *testing.T) {
	service := configv1.OpenapiUpstreamService_builder{
		SpecUrl: proto.String("http://example.com/spec.json"),
	}.Build()
	stripSecretsFromOpenapiService(service)
	assert.Equal(t, "http://example.com/spec.json", service.GetSpecUrl())
}

func TestStripSecretsFromMcpCall(t *testing.T) {
	call := configv1.MCPCallDefinition_builder{
		InputSchema: nil,
	}.Build()
	stripSecretsFromMcpCall(call)
}

func TestStripSecretsFromMcpService_StdioEnv(t *testing.T) {
	// Test environment variable stripping in McpService Stdio
	mcp := configv1.McpUpstreamService_builder{
		StdioConnection: configv1.McpStdioConnection_builder{
			Command: proto.String("python"),
			Env: map[string]*configv1.SecretValue{
				"API_KEY": configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("MY_KEY"),
				}.Build(),
				"OTHER": configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("OTHER_VAL"),
				}.Build(),
			},
		}.Build(),
	}.Build()

	stripSecretsFromMcpService(mcp)
	env := mcp.GetStdioConnection().GetEnv()
	assert.Equal(t, "MY_KEY", env["API_KEY"].GetEnvironmentVariable())

	// Test scrubbing plain text
	mcp2 := configv1.McpUpstreamService_builder{
		BundleConnection: configv1.McpBundleConnection_builder{
			BundlePath: proto.String("./bundle.zip"),
			Env: map[string]*configv1.SecretValue{
				"SECRET": configv1.SecretValue_builder{
					PlainText: proto.String("sensitive"),
				}.Build(),
			},
		}.Build(),
	}.Build()
	stripSecretsFromMcpService(mcp2)
	env2 := mcp2.GetBundleConnection().GetEnv()
	assert.False(t, env2["SECRET"].HasValue())
}

func TestScrubSecretValue(t *testing.T) {
	// Nil
	scrubSecretValue(nil)

	// Plain text
	sv := configv1.SecretValue_builder{
		PlainText: proto.String("secret"),
	}.Build()
	scrubSecretValue(sv)
	assert.False(t, sv.HasValue())

	// Env var
	sv2 := configv1.SecretValue_builder{
		EnvironmentVariable: proto.String("ENV"),
	}.Build()
	scrubSecretValue(sv2)
	assert.True(t, sv2.HasValue())
}

func TestStripSecretsFromHook(t *testing.T) {
	// Nil hook
	stripSecretsFromHook(nil)

	// Webhook with secret
	hook := configv1.CallHook_builder{
		Webhook: configv1.WebhookConfig_builder{
			Url:           "http://example.com",
			WebhookSecret: "secret",
		}.Build(),
	}.Build()
	stripSecretsFromHook(hook)
	assert.Equal(t, "", hook.GetWebhook().GetWebhookSecret())
}

func TestStripSecretsFromCacheConfig(t *testing.T) {
	// Nil
	stripSecretsFromCacheConfig(nil)

	// OpenAI ApiKey
	cache := configv1.CacheConfig_builder{
		SemanticConfig: configv1.SemanticCacheConfig_builder{
			Openai: configv1.OpenAIEmbeddingProviderConfig_builder{
				ApiKey: configv1.SecretValue_builder{
					PlainText: proto.String("key"),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()
	stripSecretsFromCacheConfig(cache)
	assert.False(t, cache.GetSemanticConfig().GetOpenai().GetApiKey().HasValue())

	// Deprecated ApiKey
	cache2 := configv1.CacheConfig_builder{
		SemanticConfig: configv1.SemanticCacheConfig_builder{
			ApiKey: configv1.SecretValue_builder{
				PlainText: proto.String("key"),
			}.Build(),
		}.Build(),
	}.Build()
	stripSecretsFromCacheConfig(cache2)
	assert.False(t, cache2.GetSemanticConfig().GetApiKey().HasValue())
}

func TestStripSecretsFromCalls(t *testing.T) {
	// HTTP Call
	httpCall := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Secret: configv1.SecretValue_builder{
					PlainText: proto.String("s"),
				}.Build(),
			}.Build(),
		},
	}.Build()
	stripSecretsFromHTTPCall(httpCall)
	assert.False(t, httpCall.GetParameters()[0].GetSecret().HasValue())

	// Websocket Call
	wsCall := configv1.WebsocketCallDefinition_builder{
		Parameters: []*configv1.WebsocketParameterMapping{
			configv1.WebsocketParameterMapping_builder{
				Secret: configv1.SecretValue_builder{
					PlainText: proto.String("s"),
				}.Build(),
			}.Build(),
		},
	}.Build()
	stripSecretsFromWebsocketCall(wsCall)
	assert.False(t, wsCall.GetParameters()[0].GetSecret().HasValue())

	// Webrtc Call
	webrtcCall := configv1.WebrtcCallDefinition_builder{
		Parameters: []*configv1.WebrtcParameterMapping{
			configv1.WebrtcParameterMapping_builder{
				Secret: configv1.SecretValue_builder{
					PlainText: proto.String("s"),
				}.Build(),
			}.Build(),
		},
	}.Build()
	stripSecretsFromWebrtcCall(webrtcCall)
	assert.False(t, webrtcCall.GetParameters()[0].GetSecret().HasValue())

	// CommandLine Call
	cmdCall := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Secret: configv1.SecretValue_builder{
					PlainText: proto.String("s"),
				}.Build(),
			}.Build(),
		},
	}.Build()
	stripSecretsFromCommandLineCall(cmdCall)
	assert.False(t, cmdCall.GetParameters()[0].GetSecret().HasValue())
}

func TestHydrateSecretValue_Internal(t *testing.T) {
	secrets := map[string]*configv1.SecretValue{
        "API_KEY": configv1.SecretValue_builder{
            PlainText: proto.String("12345"),
        }.Build(),
    }

	// Test hydrateSecretValue with non-env var
    plainSecret := configv1.SecretValue_builder{PlainText: proto.String("plain")}.Build()
    hydrateSecretValue(plainSecret, secrets) // Should do nothing
	assert.Equal(t, "plain", plainSecret.GetPlainText())

    // Test hydrateSecretValue with missing secret key
    missingSecret := configv1.SecretValue_builder{EnvironmentVariable: proto.String("MISSING")}.Build()
    hydrateSecretValue(missingSecret, secrets) // Should do nothing
	assert.Equal(t, "MISSING", missingSecret.GetEnvironmentVariable())
}

// Tests migrated from secrets_test.go

func TestStripSecretsFromService(t *testing.T) {
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		UpstreamAuth: configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				ParamName: proto.String("X-API-Key"),
				Value: configv1.SecretValue_builder{
					PlainText: proto.String("secret-key"),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()

	StripSecretsFromService(svc)

	assert.NotNil(t, svc.GetUpstreamAuth())
	assert.NotNil(t, svc.GetUpstreamAuth().GetApiKey())
	assert.NotNil(t, svc.GetUpstreamAuth().GetApiKey().GetValue())
	assert.False(t, svc.GetUpstreamAuth().GetApiKey().GetValue().HasValue(), "Plain text secret should be cleared")
}

func TestStripSecretsFromProfile(t *testing.T) {
	profile := configv1.ProfileDefinition_builder{
		Name: proto.String("test-profile"),
		Secrets: map[string]*configv1.SecretValue{
			"TEST_SECRET": configv1.SecretValue_builder{PlainText: proto.String("secret-value")}.Build(),
		},
	}.Build()

	StripSecretsFromProfile(profile)

	secret := profile.GetSecrets()["TEST_SECRET"]
	assert.NotNil(t, secret)
	assert.False(t, secret.HasValue(), "Plain text secret should be cleared")
}

func TestStripSecretsFromCollection(t *testing.T) {
	collection := configv1.Collection_builder{
		Name: proto.String("test-collection"),
		Services: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: proto.String("svc1"),
				UpstreamAuth: configv1.Authentication_builder{
					BasicAuth: configv1.BasicAuth_builder{
						Username: proto.String("user"),
						Password: configv1.SecretValue_builder{
							PlainText: proto.String("secret-password"),
						}.Build(),
					}.Build(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	StripSecretsFromCollection(collection)

	svc := collection.GetServices()[0]
	assert.NotNil(t, svc.GetUpstreamAuth())
	assert.False(t, svc.GetUpstreamAuth().GetBasicAuth().GetPassword().HasValue(), "Plain text secret should be cleared")
}

func TestHydrateSecretsInService(t *testing.T) {
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		UpstreamAuth: configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				ParamName: proto.String("X-API-Key"),
				Value: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("API_KEY_VAR"),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()

	secrets := map[string]*configv1.SecretValue{
		"API_KEY_VAR": configv1.SecretValue_builder{PlainText: proto.String("resolved-secret")}.Build(),
	}

	HydrateSecretsInService(svc, secrets)

	assert.True(t, svc.GetUpstreamAuth().GetApiKey().GetValue().HasPlainText())
	val := svc.GetUpstreamAuth().GetApiKey().GetValue().GetPlainText()
	assert.Equal(t, "resolved-secret", val)
}

func TestHydrateSecretsInService_HttpService(t *testing.T) {
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-http-service"),
		HttpService: configv1.HttpUpstreamService_builder{
			Calls: map[string]*configv1.HttpCallDefinition{
				"call1": configv1.HttpCallDefinition_builder{
					Parameters: []*configv1.HttpParameterMapping{
						configv1.HttpParameterMapping_builder{
							Schema: configv1.ParameterSchema_builder{Name: proto.String("apiKey")}.Build(),
							Secret: configv1.SecretValue_builder{
								EnvironmentVariable: proto.String("API_KEY_VAR"),
							}.Build(),
						}.Build(),
					},
				}.Build(),
			},
		}.Build(),
	}.Build()

	secrets := map[string]*configv1.SecretValue{
		"API_KEY_VAR": configv1.SecretValue_builder{PlainText: proto.String("resolved-secret")}.Build(),
	}

	HydrateSecretsInService(svc, secrets)

	httpSvc := svc.GetHttpService()
	assert.NotNil(t, httpSvc)
	call := httpSvc.GetCalls()["call1"]
	assert.NotNil(t, call)
	param := call.GetParameters()[0]
	assert.NotNil(t, param.GetSecret())

	assert.True(t, param.GetSecret().HasPlainText())
	assert.Equal(t, "resolved-secret", param.GetSecret().GetPlainText())
}

func TestHydrateSecretsInService_WebsocketService(t *testing.T) {
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-ws-service"),
		WebsocketService: configv1.WebsocketUpstreamService_builder{
			Calls: map[string]*configv1.WebsocketCallDefinition{
				"call1": configv1.WebsocketCallDefinition_builder{
					Parameters: []*configv1.WebsocketParameterMapping{
						configv1.WebsocketParameterMapping_builder{
							Schema: configv1.ParameterSchema_builder{Name: proto.String("token")}.Build(),
							Secret: configv1.SecretValue_builder{
								EnvironmentVariable: proto.String("WS_TOKEN"),
							}.Build(),
						}.Build(),
					},
				}.Build(),
			},
		}.Build(),
	}.Build()

	secrets := map[string]*configv1.SecretValue{
		"WS_TOKEN": configv1.SecretValue_builder{PlainText: proto.String("resolved-token")}.Build(),
	}

	HydrateSecretsInService(svc, secrets)

	wsSvc := svc.GetWebsocketService()
	assert.NotNil(t, wsSvc)
	call := wsSvc.GetCalls()["call1"]
	assert.NotNil(t, call)
	param := call.GetParameters()[0]
	assert.NotNil(t, param.GetSecret())

	assert.True(t, param.GetSecret().HasPlainText())
	assert.Equal(t, "resolved-token", param.GetSecret().GetPlainText())
}

func TestHydrateSecretsInService_WebrtcService(t *testing.T) {
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-webrtc-service"),
		WebrtcService: configv1.WebrtcUpstreamService_builder{
			Calls: map[string]*configv1.WebrtcCallDefinition{
				"call1": configv1.WebrtcCallDefinition_builder{
					Parameters: []*configv1.WebrtcParameterMapping{
						configv1.WebrtcParameterMapping_builder{
							Schema: configv1.ParameterSchema_builder{Name: proto.String("secret")}.Build(),
							Secret: configv1.SecretValue_builder{
								EnvironmentVariable: proto.String("RTC_SECRET"),
							}.Build(),
						}.Build(),
					},
				}.Build(),
			},
		}.Build(),
	}.Build()

	secrets := map[string]*configv1.SecretValue{
		"RTC_SECRET": configv1.SecretValue_builder{PlainText: proto.String("resolved-secret")}.Build(),
	}

	HydrateSecretsInService(svc, secrets)

	rtcSvc := svc.GetWebrtcService()
	assert.NotNil(t, rtcSvc)
	call := rtcSvc.GetCalls()["call1"]
	assert.NotNil(t, call)
	param := call.GetParameters()[0]
	assert.NotNil(t, param.GetSecret())

	assert.True(t, param.GetSecret().HasPlainText())
	assert.Equal(t, "resolved-secret", param.GetSecret().GetPlainText())
}

func TestHydrateSecretsInService_CommandLineService(t *testing.T) {
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-cmd-service"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Env: map[string]*configv1.SecretValue{
				"ENV_VAR": configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("SECRET_ENV"),
				}.Build(),
			},
			ContainerEnvironment: configv1.ContainerEnvironment_builder{
				Env: map[string]*configv1.SecretValue{
					"CONTAINER_VAR": configv1.SecretValue_builder{
						EnvironmentVariable: proto.String("SECRET_CONTAINER"),
					}.Build(),
				},
			}.Build(),
		}.Build(),
	}.Build()

	secrets := map[string]*configv1.SecretValue{
		"SECRET_ENV":       configv1.SecretValue_builder{PlainText: proto.String("env-value")}.Build(),
		"SECRET_CONTAINER": configv1.SecretValue_builder{PlainText: proto.String("container-value")}.Build(),
	}

	HydrateSecretsInService(svc, secrets)

	cmdSvc := svc.GetCommandLineService()
	assert.NotNil(t, cmdSvc)

	envVal := cmdSvc.GetEnv()["ENV_VAR"]
	if assert.True(t, envVal.HasPlainText()) {
		assert.Equal(t, "env-value", envVal.GetPlainText())
	}

	containerVal := cmdSvc.GetContainerEnvironment().GetEnv()["CONTAINER_VAR"]
	if assert.True(t, containerVal.HasPlainText()) {
		assert.Equal(t, "container-value", containerVal.GetPlainText())
	}
}

func TestHydrateSecretsInService_McpService(t *testing.T) {
	// Stdio
	svcStdio := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-mcp-stdio"),
		McpService: configv1.McpUpstreamService_builder{
			StdioConnection: configv1.McpStdioConnection_builder{
				Env: map[string]*configv1.SecretValue{
					"STDIO_VAR": configv1.SecretValue_builder{
						EnvironmentVariable: proto.String("SECRET_STDIO"),
					}.Build(),
				},
			}.Build(),
		}.Build(),
	}.Build()

	// Bundle
	svcBundle := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-mcp-bundle"),
		McpService: configv1.McpUpstreamService_builder{
			BundleConnection: configv1.McpBundleConnection_builder{
				Env: map[string]*configv1.SecretValue{
					"BUNDLE_VAR": configv1.SecretValue_builder{
						EnvironmentVariable: proto.String("SECRET_BUNDLE"),
					}.Build(),
				},
			}.Build(),
		}.Build(),
	}.Build()

	secrets := map[string]*configv1.SecretValue{
		"SECRET_STDIO":  configv1.SecretValue_builder{PlainText: proto.String("stdio-value")}.Build(),
		"SECRET_BUNDLE": configv1.SecretValue_builder{PlainText: proto.String("bundle-value")}.Build(),
	}

	HydrateSecretsInService(svcStdio, secrets)
	HydrateSecretsInService(svcBundle, secrets)

	stdioEnvVal := svcStdio.GetMcpService().GetStdioConnection().GetEnv()["STDIO_VAR"]
	if assert.True(t, stdioEnvVal.HasPlainText()) {
		assert.Equal(t, "stdio-value", stdioEnvVal.GetPlainText())
	}

	bundleEnvVal := svcBundle.GetMcpService().GetBundleConnection().GetEnv()["BUNDLE_VAR"]
	if assert.True(t, bundleEnvVal.HasPlainText()) {
		assert.Equal(t, "bundle-value", bundleEnvVal.GetPlainText())
	}
}

func TestStripSecretsFromAllServices(t *testing.T) {
	// Setup services with secrets
	services := []*configv1.UpstreamServiceConfig{
		configv1.UpstreamServiceConfig_builder{
			Name: proto.String("http"),
			HttpService: configv1.HttpUpstreamService_builder{
				Calls: map[string]*configv1.HttpCallDefinition{
					"c": configv1.HttpCallDefinition_builder{
						Parameters: []*configv1.HttpParameterMapping{
							configv1.HttpParameterMapping_builder{
								Secret: configv1.SecretValue_builder{PlainText: proto.String("s")}.Build(),
							}.Build(),
						},
					}.Build(),
				},
			}.Build(),
		}.Build(),
		configv1.UpstreamServiceConfig_builder{
			Name: proto.String("ws"),
			WebsocketService: configv1.WebsocketUpstreamService_builder{
				Calls: map[string]*configv1.WebsocketCallDefinition{
					"c": configv1.WebsocketCallDefinition_builder{
						Parameters: []*configv1.WebsocketParameterMapping{
							configv1.WebsocketParameterMapping_builder{
								Secret: configv1.SecretValue_builder{PlainText: proto.String("s")}.Build(),
							}.Build(),
						},
					}.Build(),
				},
			}.Build(),
		}.Build(),
		configv1.UpstreamServiceConfig_builder{
			Name: proto.String("webrtc"),
			WebrtcService: configv1.WebrtcUpstreamService_builder{
				Calls: map[string]*configv1.WebrtcCallDefinition{
					"c": configv1.WebrtcCallDefinition_builder{
						Parameters: []*configv1.WebrtcParameterMapping{
							configv1.WebrtcParameterMapping_builder{
								Secret: configv1.SecretValue_builder{PlainText: proto.String("s")}.Build(),
							}.Build(),
						},
					}.Build(),
				},
			}.Build(),
		}.Build(),
		configv1.UpstreamServiceConfig_builder{
			Name: proto.String("mcp"),
			McpService: configv1.McpUpstreamService_builder{
				StdioConnection: configv1.McpStdioConnection_builder{
					Env: map[string]*configv1.SecretValue{"k": configv1.SecretValue_builder{PlainText: proto.String("s")}.Build()},
				}.Build(),
			}.Build(),
		}.Build(),
		configv1.UpstreamServiceConfig_builder{
			Name: proto.String("vector"),
			VectorService: configv1.VectorUpstreamService_builder{
				Pinecone: configv1.PineconeVectorDB_builder{ApiKey: proto.String("s")}.Build(),
			}.Build(),
		}.Build(),
		configv1.UpstreamServiceConfig_builder{
			Name: proto.String("fs"),
			FilesystemService: configv1.FilesystemUpstreamService_builder{
				S3: configv1.S3Fs_builder{SecretAccessKey: proto.String("s")}.Build(),
			}.Build(),
		}.Build(),
	}

	for _, svc := range services {
		StripSecretsFromService(svc)
	}

	// Verify secrets are gone
	assert.False(t, services[0].GetHttpService().GetCalls()["c"].GetParameters()[0].GetSecret().HasValue())
	assert.False(t, services[1].GetWebsocketService().GetCalls()["c"].GetParameters()[0].GetSecret().HasValue())
	assert.False(t, services[2].GetWebrtcService().GetCalls()["c"].GetParameters()[0].GetSecret().HasValue())
	assert.False(t, services[3].GetMcpService().GetStdioConnection().GetEnv()["k"].HasValue())
	assert.Equal(t, "", services[4].GetVectorService().GetPinecone().GetApiKey())
	assert.Equal(t, "", services[5].GetFilesystemService().GetS3().GetSecretAccessKey())
}

func TestStripSecretsFromGrpcAndOpenapi(t *testing.T) {
	// These currently don't do much, but we test for coverage
	svcGrpc := configv1.UpstreamServiceConfig_builder{
		Id: proto.String("grpc"),
		GrpcService: configv1.GrpcUpstreamService_builder{}.Build(),
	}.Build()
	StripSecretsFromService(svcGrpc)

	svcOpenapi := configv1.UpstreamServiceConfig_builder{
		Id: proto.String("openapi"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{}.Build(),
	}.Build()
	StripSecretsFromService(svcOpenapi)
}

func TestStripSecretsFromHookAndCache(t *testing.T) {
	svc := configv1.UpstreamServiceConfig_builder{
		Id: proto.String("hook-cache"),
		PreCallHooks: []*configv1.CallHook{
			configv1.CallHook_builder{
				Webhook: configv1.WebhookConfig_builder{WebhookSecret: "secret"}.Build(),
			}.Build(),
		},
		Cache: configv1.CacheConfig_builder{
			SemanticConfig: configv1.SemanticCacheConfig_builder{
				ApiKey: configv1.SecretValue_builder{PlainText: proto.String("s")}.Build(),
				Openai: configv1.OpenAIEmbeddingProviderConfig_builder{
					ApiKey: configv1.SecretValue_builder{PlainText: proto.String("s")}.Build(),
				}.Build(),
			}.Build(),
		}.Build(),
	}.Build()

	StripSecretsFromService(svc)

	assert.Equal(t, "", svc.GetPreCallHooks()[0].GetWebhook().GetWebhookSecret())
	assert.False(t, svc.GetCache().GetSemanticConfig().GetApiKey().HasValue())
	assert.False(t, svc.GetCache().GetSemanticConfig().GetOpenai().GetApiKey().HasValue())
}

func TestHydrateSecretsInAuth_Extended(t *testing.T) {
	auth := configv1.Authentication_builder{
		BearerToken: configv1.BearerTokenAuth_builder{
			Token: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("TOKEN_VAR"),
			}.Build(),
		}.Build(),
	}.Build()

	secrets := map[string]*configv1.SecretValue{
		"TOKEN_VAR": configv1.SecretValue_builder{PlainText: proto.String("token")}.Build(),
	}

	hydrateSecretsInAuth(auth, secrets)
	assert.Equal(t, "token", auth.GetBearerToken().GetToken().GetPlainText())

	authBasic := configv1.Authentication_builder{
		BasicAuth: configv1.BasicAuth_builder{
			Password: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("PASS_VAR"),
			}.Build(),
		}.Build(),
	}.Build()
	secrets["PASS_VAR"] = configv1.SecretValue_builder{PlainText: proto.String("pass")}.Build()
	hydrateSecretsInAuth(authBasic, secrets)
	assert.Equal(t, "pass", authBasic.GetBasicAuth().GetPassword().GetPlainText())

	authOauth := configv1.Authentication_builder{
		Oauth2: configv1.OAuth2Auth_builder{
			ClientId: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("ID_VAR"),
			}.Build(),
			ClientSecret: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("SECRET_VAR"),
			}.Build(),
		}.Build(),
	}.Build()
	secrets["ID_VAR"] = configv1.SecretValue_builder{PlainText: proto.String("id")}.Build()
	secrets["SECRET_VAR"] = configv1.SecretValue_builder{PlainText: proto.String("secret")}.Build()
	hydrateSecretsInAuth(authOauth, secrets)
	assert.Equal(t, "id", authOauth.GetOauth2().GetClientId().GetPlainText())
	assert.Equal(t, "secret", authOauth.GetOauth2().GetClientSecret().GetPlainText())
}

func TestCoverageForEmptyStripFunctions(t *testing.T) {
	// Call empty functions to ensure coverage
	stripSecretsFromGrpcService(nil)
	stripSecretsFromOpenapiService(nil)
	stripSecretsFromMcpCall(nil)
}

// TestStripSecretsFromService_NilBranches ensures that our recursive stripping
// functions safely handle nil or empty structures without panicking, which is a
// required property of our configuration sanitization contract.
func TestStripSecretsFromService_NilBranches(t *testing.T) {
	tests := []struct {
		name     string
		sanitize func()
	}{
		{
			name: "CommandLineService with nil",
			sanitize: func() { stripSecretsFromCommandLineService(nil) },
		},
		{
			name: "HTTPService with nil",
			sanitize: func() { stripSecretsFromHTTPService(nil) },
		},
		{
			name: "McpService with nil",
			sanitize: func() { stripSecretsFromMcpService(nil) },
		},
		{
			name: "FilesystemService with nil",
			sanitize: func() { stripSecretsFromFilesystemService(nil) },
		},
		{
			name: "VectorService with nil",
			sanitize: func() { stripSecretsFromVectorService(nil) },
		},
		{
			name: "WebsocketService with nil",
			sanitize: func() { stripSecretsFromWebsocketService(nil) },
		},
		{
			name: "WebrtcService with nil",
			sanitize: func() { stripSecretsFromWebrtcService(nil) },
		},
		{
			name: "Hook with nil",
			sanitize: func() { stripSecretsFromHook(nil) },
		},
		{
			name: "CommandLineCall with nil",
			sanitize: func() { stripSecretsFromCommandLineCall(nil) },
		},
		{
			name: "HTTPCall with nil",
			sanitize: func() { stripSecretsFromHTTPCall(nil) },
		},
		{
			name: "WebsocketCall with nil",
			sanitize: func() { stripSecretsFromWebsocketCall(nil) },
		},
		{
			name: "WebrtcCall with nil",
			sanitize: func() { stripSecretsFromWebrtcCall(nil) },
		},
		{
			name: "SecretValue with nil",
			sanitize: func() { scrubSecretValue(nil) },
		},
		{
			name: "Profile with nil",
			sanitize: func() { StripSecretsFromProfile(nil) },
		},
		{
			name: "Collection with nil",
			sanitize: func() { StripSecretsFromCollection(nil) },
		},
		{
			name: "Service with nil config",
			sanitize: func() { StripSecretsFromService(nil) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The assertion is that these do not panic.
			requireNotPanic(t, tt.sanitize)
		})
	}
}

// requireNotPanic is a helper to ensure a function executes without panicking.
func requireNotPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked: %v", r)
		}
	}()
	f()
}

// TestStripSecretsFromService_VectorAndFS explicitly tests the branches
// for VectorDb and Filesystem configurations to ensure credentials are removed.
func TestStripSecretsFromService_VectorAndFS(t *testing.T) {
	// 1. Test Milvus VectorDB
	sMilvus := configv1.VectorUpstreamService_builder{
		Milvus: configv1.MilvusVectorDB_builder{
			ApiKey:   proto.String("test-api-key"),
			Password: proto.String("test-password"),
		}.Build(),
	}.Build()

	// 1b. Test with raw struct literal for Milvus VectorDB
	// sMilvus := &configv1.VectorUpstreamService{
	// 	VectorDbType: &configv1.VectorUpstreamService_Milvus{
	// 		Milvus: &configv1.MilvusVectorDB{
	// 			ApiKey:   "test-api-key",
	// 			Password: "test-password",
	// 		},
	// 	},
	// }

	stripSecretsFromVectorService(sMilvus)

	if sMilvus.GetMilvus().GetApiKey() != "" {
		t.Errorf("expected milvus API key to be stripped, got %q", sMilvus.GetMilvus().GetApiKey())
	}
	if sMilvus.GetMilvus().GetPassword() != "" {
		t.Errorf("expected milvus password to be stripped, got %q", sMilvus.GetMilvus().GetPassword())
	}

	// 2. Test SFTP Filesystem
	sSftp := configv1.FilesystemUpstreamService_builder{
		Sftp: configv1.SftpFs_builder{
			Password: proto.String("test-password"),
		}.Build(),
	}.Build()

	stripSecretsFromFilesystemService(sSftp)

	if sSftp.GetSftp().GetPassword() != "" {
		t.Errorf("expected sftp password to be stripped, got %q", sSftp.GetSftp().GetPassword())
	}
}

// TestEmptyServiceSignatures ensures that services with empty/no-op secret stripping
// implementations still fulfill the contract without error.
func TestEmptyServiceSignatures(t *testing.T) {
	t.Run("GrpcService", func(t *testing.T) {
		sGrpc := configv1.GrpcUpstreamService_builder{}.Build()
		requireNotPanic(t, func() { stripSecretsFromGrpcService(sGrpc) })
		requireNotPanic(t, func() { stripSecretsFromGrpcService(nil) })
	})

	t.Run("OpenapiService", func(t *testing.T) {
		sOpenapi := configv1.OpenapiUpstreamService_builder{}.Build()
		requireNotPanic(t, func() { stripSecretsFromOpenapiService(sOpenapi) })
		requireNotPanic(t, func() { stripSecretsFromOpenapiService(nil) })
	})

	t.Run("McpCall", func(t *testing.T) {
		sMcpCall := configv1.MCPCallDefinition_builder{}.Build()
		requireNotPanic(t, func() { stripSecretsFromMcpCall(sMcpCall) })
		requireNotPanic(t, func() { stripSecretsFromMcpCall(nil) })
	})
}
