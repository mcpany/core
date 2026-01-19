// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestStripSecretsFromGrpcService(t *testing.T) {
	service := &configv1.GrpcUpstreamService{
		Address: proto.String("localhost:50051"),
		Calls: map[string]*configv1.GrpcCallDefinition{
			"call1": {
				InputSchema:  nil,
				OutputSchema: nil,
			},
		},
	}
	// Currently stripSecretsFromGrpcService is a placeholder or does minimal work
	// But we need to call it to get coverage.

	stripSecretsFromGrpcService(service)
	assert.Equal(t, "localhost:50051", service.GetAddress())
}

func TestStripSecretsFromOpenapiService(t *testing.T) {
	service := &configv1.OpenapiUpstreamService{
		SpecSource: &configv1.OpenapiUpstreamService_SpecUrl{
			SpecUrl: "http://example.com/spec.json",
		},
	}
	stripSecretsFromOpenapiService(service)
	assert.Equal(t, "http://example.com/spec.json", service.GetSpecUrl())
}

func TestStripSecretsFromMcpCall(t *testing.T) {
	call := &configv1.MCPCallDefinition{
		InputSchema: nil,
	}
	stripSecretsFromMcpCall(call)
}

func TestStripSecretsFromMcpService_StdioEnv(t *testing.T) {
	// Test environment variable stripping in McpService Stdio
	mcp := &configv1.McpUpstreamService{
		ConnectionType: &configv1.McpUpstreamService_StdioConnection{
			StdioConnection: &configv1.McpStdioConnection{
				Command: proto.String("python"),
				Env: map[string]*configv1.SecretValue{
					"API_KEY": {
						Value: &configv1.SecretValue_EnvironmentVariable{
							EnvironmentVariable: "MY_KEY",
						},
					},
					"OTHER": {
						Value: &configv1.SecretValue_EnvironmentVariable{
							EnvironmentVariable: "OTHER_VAL",
						},
					},
				},
			},
		},
	}

	stripSecretsFromMcpService(mcp)
	env := mcp.GetStdioConnection().GetEnv()
	// Note: scrubSecretValue removes the value if it is PlainText,
	// but here we have EnvironmentVariable, so it should stay?
	// scrubSecretValue logic:
	// 	if _, ok := sv.Value.(*configv1.SecretValue_PlainText); ok {
	//		sv.Value = nil
	//	}
	// So Env var should be preserved!
	assert.Equal(t, "MY_KEY", env["API_KEY"].GetEnvironmentVariable())

	// Test scrubbing plain text
	mcp2 := &configv1.McpUpstreamService{
		ConnectionType: &configv1.McpUpstreamService_BundleConnection{
			BundleConnection: &configv1.McpBundleConnection{
				BundlePath: proto.String("./bundle.zip"),
				Env: map[string]*configv1.SecretValue{
					"SECRET": {
						Value: &configv1.SecretValue_PlainText{
							PlainText: "sensitive",
						},
					},
				},
			},
		},
	}
	stripSecretsFromMcpService(mcp2)
	env2 := mcp2.GetBundleConnection().GetEnv()
	assert.Nil(t, env2["SECRET"].Value)
}

func TestScrubSecretValue(t *testing.T) {
	// Nil
	scrubSecretValue(nil)

	// Plain text
	sv := &configv1.SecretValue{
		Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
	}
	scrubSecretValue(sv)
	assert.Nil(t, sv.Value)

	// Env var
	sv2 := &configv1.SecretValue{
		Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "ENV"},
	}
	scrubSecretValue(sv2)
	assert.NotNil(t, sv2.Value)
}

func TestStripSecretsFromHook(t *testing.T) {
	// Nil hook
	stripSecretsFromHook(nil)

	// Webhook with secret
	hook := &configv1.CallHook{
		HookConfig: &configv1.CallHook_Webhook{
			Webhook: &configv1.WebhookConfig{
				Url:           "http://example.com",
				WebhookSecret: "secret",
			},
		},
	}
	stripSecretsFromHook(hook)
	assert.Equal(t, "", hook.GetWebhook().GetWebhookSecret())
}

func TestStripSecretsFromCacheConfig(t *testing.T) {
	// Nil
	stripSecretsFromCacheConfig(nil)

	// OpenAI ApiKey
	cache := &configv1.CacheConfig{
		SemanticConfig: &configv1.SemanticCacheConfig{
			ProviderConfig: &configv1.SemanticCacheConfig_Openai{
				Openai: &configv1.OpenAIEmbeddingProviderConfig{
					ApiKey: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{PlainText: "key"},
					},
				},
			},
		},
	}
	stripSecretsFromCacheConfig(cache)
	assert.Nil(t, cache.SemanticConfig.GetOpenai().ApiKey.Value)

	// Deprecated ApiKey
	cache2 := &configv1.CacheConfig{
		SemanticConfig: &configv1.SemanticCacheConfig{
			ApiKey: &configv1.SecretValue{
				Value: &configv1.SecretValue_PlainText{PlainText: "key"},
			},
		},
	}
	stripSecretsFromCacheConfig(cache2)
	assert.Nil(t, cache2.SemanticConfig.ApiKey.Value)
}

func TestStripSecretsFromCalls(t *testing.T) {
	// HTTP Call
	httpCall := &configv1.HttpCallDefinition{
		Parameters: []*configv1.HttpParameterMapping{
			{
				Secret: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "s"},
				},
			},
		},
	}
	stripSecretsFromHTTPCall(httpCall)
	assert.Nil(t, httpCall.Parameters[0].Secret.Value)

	// Websocket Call
	wsCall := &configv1.WebsocketCallDefinition{
		Parameters: []*configv1.WebsocketParameterMapping{
			{
				Secret: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "s"},
				},
			},
		},
	}
	stripSecretsFromWebsocketCall(wsCall)
	assert.Nil(t, wsCall.Parameters[0].Secret.Value)

	// Webrtc Call
	webrtcCall := &configv1.WebrtcCallDefinition{
		Parameters: []*configv1.WebrtcParameterMapping{
			{
				Secret: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "s"},
				},
			},
		},
	}
	stripSecretsFromWebrtcCall(webrtcCall)
	assert.Nil(t, webrtcCall.Parameters[0].Secret.Value)

	// CommandLine Call
	cmdCall := &configv1.CommandLineCallDefinition{
		Parameters: []*configv1.CommandLineParameterMapping{
			{
				Secret: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "s"},
				},
			},
		},
	}
	stripSecretsFromCommandLineCall(cmdCall)
	assert.Nil(t, cmdCall.Parameters[0].Secret.Value)
}
