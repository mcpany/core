// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package lint

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLinter_Run_AllAuthTypes_PlainText(t *testing.T) {
	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("service-all-auth"),
				UpstreamAuth: configv1.Authentication_builder{
					ApiKey: configv1.APIKeyAuth_builder{
						Value: configv1.SecretValue_builder{
							PlainText: proto.String("api-key"),
						}.Build(),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("service-bearer"),
				UpstreamAuth: configv1.Authentication_builder{
					BearerToken: configv1.BearerTokenAuth_builder{
						Token: configv1.SecretValue_builder{
							PlainText: proto.String("bearer-token"),
						}.Build(),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("service-basic"),
				UpstreamAuth: configv1.Authentication_builder{
					BasicAuth: configv1.BasicAuth_builder{
						Password: configv1.SecretValue_builder{
							PlainText: proto.String("basic-password"),
						}.Build(),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("service-oauth"),
				UpstreamAuth: configv1.Authentication_builder{
					Oauth2: configv1.OAuth2Auth_builder{
						ClientSecret: configv1.SecretValue_builder{
							PlainText: proto.String("oauth-secret"),
						}.Build(),
					}.Build(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	linter := NewLinter(cfg)
	results, err := linter.Run(context.Background())
	assert.NoError(t, err)

	count := 0
	for _, r := range results {
		if r.Severity == Warning && r.Message == "Secret is stored in plain text. Use environment variables or file references for better security." {
			count++
		}
	}
	assert.Equal(t, 4, count, "Expected 4 warnings about plain text secrets")
}

func TestLinter_Run_EnvVars_PlainText(t *testing.T) {
	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("cmd-env"),
				CommandLineService: configv1.CommandLineUpstreamService_builder{
					Env: map[string]*configv1.SecretValue{
						"KEY": configv1.SecretValue_builder{PlainText: proto.String("val")}.Build(),
					},
					ContainerEnvironment: configv1.ContainerEnvironment_builder{
						Env: map[string]*configv1.SecretValue{
							"CONTAINER_KEY": configv1.SecretValue_builder{PlainText: proto.String("val")}.Build(),
						},
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("mcp-env"),
				McpService: configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{
						Env: map[string]*configv1.SecretValue{
							"STDIO_KEY": configv1.SecretValue_builder{PlainText: proto.String("val")}.Build(),
						},
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("mcp-bundle-env"),
				McpService: configv1.McpUpstreamService_builder{
					BundleConnection: configv1.McpBundleConnection_builder{
						Env: map[string]*configv1.SecretValue{
							"BUNDLE_KEY": configv1.SecretValue_builder{PlainText: proto.String("val")}.Build(),
						},
					}.Build(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	linter := NewLinter(cfg)
	results, err := linter.Run(context.Background())
	assert.NoError(t, err)

	count := 0
	for _, r := range results {
		if r.Severity == Warning && strings.Contains(r.Message, "Secret is stored in plain text") {
			count++
		}
	}
	assert.Equal(t, 4, count, "Expected 4 warnings about plain text secrets in env vars")
}

func TestLinter_Run_OtherInsecureHTTP(t *testing.T) {
	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("openapi-insecure"),
				OpenapiService: configv1.OpenapiUpstreamService_builder{
					Address: ptr("http://api.openapi.com"),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("openapi-spec-insecure"),
				OpenapiService: configv1.OpenapiUpstreamService_builder{
					SpecUrl: proto.String("http://spec.openapi.com"),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("mcp-http-insecure"),
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{
						HttpAddress: ptr("http://mcp.example.com"),
					}.Build(),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("safe-127.0.0.1"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: ptr("http://127.0.0.1:8080"),
				}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("safe-127"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: ptr("http://127.0.0.1:8080"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	linter := NewLinter(cfg)
	results, err := linter.Run(context.Background())
	assert.NoError(t, err)

	count := 0
	for _, r := range results {
		if r.Severity == Warning && strings.Contains(r.Message, "insecure HTTP connection") {
			count++
		}
	}
	assert.Equal(t, 3, count, "Expected 3 warnings about insecure HTTP")
}

func TestLinter_Run_ShellInjection_Extra(t *testing.T) {
	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("mcp-stdio-shell"),
				McpService: configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{
						Command: ptr("bash -c 'bad'"),
					}.Build(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	linter := NewLinter(cfg)
	results, err := linter.Run(context.Background())
	assert.NoError(t, err)

	found := false
	for _, r := range results {
		if r.Severity == Warning && strings.Contains(r.Message, "shell invocation") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected warning about shell injection in MCP Stdio")
}
