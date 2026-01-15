// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package lint

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestLinter_Run_AllAuthTypes_PlainText(t *testing.T) {
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: ptr("service-all-auth"),
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_ApiKey{
						ApiKey: &configv1.APIKeyAuth{
							Value: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{PlainText: "api-key"},
							},
						},
					},
				},
			},
			{
				Name: ptr("service-bearer"),
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_BearerToken{
						BearerToken: &configv1.BearerTokenAuth{
							Token: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{PlainText: "bearer-token"},
							},
						},
					},
				},
			},
			{
				Name: ptr("service-basic"),
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_BasicAuth{
						BasicAuth: &configv1.BasicAuth{
							Password: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{PlainText: "basic-password"},
							},
						},
					},
				},
			},
			{
				Name: ptr("service-oauth"),
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_Oauth2{
						Oauth2: &configv1.OAuth2Auth{
							ClientSecret: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{PlainText: "oauth-secret"},
							},
						},
					},
				},
			},
		},
	}

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
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: ptr("cmd-env"),
				ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
					CommandLineService: &configv1.CommandLineUpstreamService{
						Env: map[string]*configv1.SecretValue{
							"KEY": {Value: &configv1.SecretValue_PlainText{PlainText: "val"}},
						},
						ContainerEnvironment: &configv1.ContainerEnvironment{
							Env: map[string]*configv1.SecretValue{
								"CONTAINER_KEY": {Value: &configv1.SecretValue_PlainText{PlainText: "val"}},
							},
						},
					},
				},
			},
			{
				Name: ptr("mcp-env"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Env: map[string]*configv1.SecretValue{
									"STDIO_KEY": {Value: &configv1.SecretValue_PlainText{PlainText: "val"}},
								},
							},
						},
					},
				},
			},
			{
				Name: ptr("mcp-bundle-env"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_BundleConnection{
							BundleConnection: &configv1.McpBundleConnection{
								Env: map[string]*configv1.SecretValue{
									"BUNDLE_KEY": {Value: &configv1.SecretValue_PlainText{PlainText: "val"}},
								},
							},
						},
					},
				},
			},
		},
	}

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
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: ptr("openapi-insecure"),
				ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
					OpenapiService: &configv1.OpenapiUpstreamService{
						Address: ptr("http://api.openapi.com"),
					},
				},
			},
			{
				Name: ptr("openapi-spec-insecure"),
				ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
					OpenapiService: &configv1.OpenapiUpstreamService{
						SpecSource: &configv1.OpenapiUpstreamService_SpecUrl{
							SpecUrl: "http://spec.openapi.com",
						},
					},
				},
			},
			{
				Name: ptr("mcp-http-insecure"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_HttpConnection{
							HttpConnection: &configv1.McpStreamableHttpConnection{
								HttpAddress: ptr("http://mcp.example.com"),
							},
						},
					},
				},
			},
			{
				Name: ptr("safe-localhost"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: ptr("http://localhost:8080"),
					},
				},
			},
			{
				Name: ptr("safe-127"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: ptr("http://127.0.0.1:8080"),
					},
				},
			},
		},
	}

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
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: ptr("mcp-stdio-shell"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Command: ptr("bash -c 'bad'"),
							},
						},
					},
				},
			},
		},
	}

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
