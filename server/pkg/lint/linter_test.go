// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package lint

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
)

func ptr(s string) *string {
	return &s
}

func TestLinter_Run_PlainTextSecrets(t *testing.T) {
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: ptr("test-service"),
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_ApiKey{
						ApiKey: &configv1.APIKeyAuth{
							ParamName: ptr("key"),
							Value: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{
									PlainText: "123456",
								},
							},
						},
					},
				},
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: ptr("https://example.com"),
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
		if r.Severity == Warning && r.Message == "Secret is stored in plain text. Use environment variables or file references for better security." {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected warning about plain text secret")
}

func TestLinter_Run_ShellInjection(t *testing.T) {
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: ptr("risky-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
					CommandLineService: &configv1.CommandLineUpstreamService{
						Command: ptr("sh -c 'echo hello'"),
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
		if r.Severity == Warning && r.Path == "command" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected warning about shell injection")
}

func TestLinter_Run_InsecureHTTP(t *testing.T) {
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: ptr("insecure-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: ptr("http://api.example.com"),
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
		if r.Severity == Warning && strings.Contains(r.Message, "insecure HTTP connection") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected warning about insecure HTTP")
}

func TestLinter_Run_CacheTTL(t *testing.T) {
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: ptr("cache-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: ptr("https://api.example.com"),
					},
				},
				Cache: &configv1.CacheConfig{
					Ttl: &durationpb.Duration{Seconds: 0},
				},
			},
		},
	}

	linter := NewLinter(cfg)
	results, err := linter.Run(context.Background())
	assert.NoError(t, err)

	found := false
	for _, r := range results {
		if r.Severity == Info && strings.Contains(r.Message, "Cache is configured but has 0 TTL") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected info about 0 TTL cache")
}
