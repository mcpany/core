// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestValidate_UpstreamApiKeyMissingValue(t *testing.T) {
	ctx := context.Background()

	// Upstream Service with API Key auth that has verification_value but NO value.
	// This should be invalid for an UPSTREAM service (outgoing).
	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("test-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String("http://example.com"),
					},
				},
				UpstreamAuth: &configv1.Authentication{
					AuthMethod: &configv1.Authentication_ApiKey{
						ApiKey: &configv1.APIKeyAuth{
							ParamName:         proto.String("X-API-Key"),
							VerificationValue: proto.String("some-value"), // Should not be enough for upstream
						},
					},
				},
			},
		},
	}

	errs := Validate(ctx, config, Server)

	if len(errs) == 0 {
		t.Fatal("Expected validation error for upstream API key missing 'value', but got none")
	}

	assert.Contains(t, errs[0].Err.Error(), "api key 'value' is missing")
}
