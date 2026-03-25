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

func TestLinter_Run_PlainTextSecrets(t *testing.T) {
	sVal := configv1.SecretValue_builder{
		PlainText: proto.String("123"),
	}.Build()

	auth := configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			ParamName: ptr("key"),
			Value:     sVal,
		}.Build(),
	}.Build()

	httpSvc := configv1.HttpUpstreamService_builder{
		Address: ptr("https://example.com"),
	}.Build()

	svc := configv1.UpstreamServiceConfig_builder{
		Id:           ptr("test-service"),
		UpstreamAuth: auth,
		HttpService:  httpSvc,
	}.Build()

	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
	}.Build()

	linter := NewLinter(cfg)
	results, err := linter.Run(context.Background())
	assert.NoError(t, err)

	found := false
	msg := "Secret is stored in plain text. Use " +
		"env vars or file references for better security."
	for _, r := range results {
		if r.Severity == Warning && r.Message == msg {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected warning about plain text secret")
}

func TestLinter_Run_ShellInjection(t *testing.T) {
	cmdSvc := configv1.CommandLineUpstreamService_builder{
		Command: ptr("sh -c 'echo hello'"),
	}.Build()

	svc := configv1.UpstreamServiceConfig_builder{
		Id:                 ptr("risky-service"),
		CommandLineService: cmdSvc,
	}.Build()

	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
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

func TestLinter_Run_InsecureHTTP(t *testing.T) {
	httpSvc := configv1.HttpUpstreamService_builder{
		Address: ptr("http://api.example.com"),
	}.Build()

	svc := configv1.UpstreamServiceConfig_builder{
		Id:          ptr("insecure-service"),
		HttpService: httpSvc,
	}.Build()

	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
	}.Build()

	linter := NewLinter(cfg)
	results, err := linter.Run(context.Background())
	assert.NoError(t, err)

	found := false
	for _, r := range results {
		if r.Severity == Warning &&
			strings.Contains(r.Message, "insecure HTTP") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected warning about insecure HTTP")
}

func TestLinter_Run_CacheTTL(t *testing.T) {
	httpSvc := configv1.HttpUpstreamService_builder{
		Address: ptr("https://api.example.com"),
	}.Build()

	cacheCfg := configv1.CacheConfig_builder{
		Ttl: &durationpb.Duration{Seconds: 0},
	}.Build()

	svc := configv1.UpstreamServiceConfig_builder{
		Id:          ptr("cache-service"),
		HttpService: httpSvc,
		Cache:       cacheCfg,
	}.Build()

	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
	}.Build()

	linter := NewLinter(cfg)
	results, err := linter.Run(context.Background())
	assert.NoError(t, err)

	found := false
	for _, r := range results {
		if r.Severity == Info &&
			strings.Contains(r.Message,
				"Cache is configured but has 0 TTL") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected info about 0 TTL cache")
}
