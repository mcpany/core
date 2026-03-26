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
	"google.golang.org/protobuf/types/known/durationpb"
)

func ptr(s string) *string {
	return &s
}

// TestLinter_Run_PlainTextSecrets ...
// Summary: TestLinter_Run_PlainTextSecrets
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("test-service"),
				UpstreamAuth: configv1.Authentication_builder{
					ApiKey: configv1.APIKeyAuth_builder{
						ParamName: ptr("key"),
						Value: configv1.SecretValue_builder{
							PlainText: proto.String("123456"),
						}.Build(),
					}.Build(),
				}.Build(),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: ptr("https://example.com"),
				}.Build(),
			}.Build(),
		},
	}.Build()

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

// TestLinter_Run_ShellInjection ...
// Summary: TestLinter_Run_ShellInjection
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("risky-service"),
				CommandLineService: configv1.CommandLineUpstreamService_builder{
					Command: ptr("sh -c 'echo hello'"),
				}.Build(),
			}.Build(),
		},
	}.Build()

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

// TestLinter_Run_InsecureHTTP ...
// Summary: TestLinter_Run_InsecureHTTP
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("insecure-service"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: ptr("http://api.example.com"),
				}.Build(),
			}.Build(),
		},
	}.Build()

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

// TestLinter_Run_CacheTTL ...
// Summary: TestLinter_Run_CacheTTL
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Id: ptr("cache-service"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: ptr("https://api.example.com"),
				}.Build(),
				Cache: configv1.CacheConfig_builder{
					Ttl: &durationpb.Duration{Seconds: 0},
				}.Build(),
			}.Build(),
		},
	}.Build()

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
