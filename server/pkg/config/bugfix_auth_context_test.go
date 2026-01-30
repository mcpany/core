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
	// Upstream Service with API Key auth that has verification_value but NO value.
	// This should be invalid for an UPSTREAM service (outgoing).
	httpSvc := configv1.HttpUpstreamService_builder{
		Address: proto.String("http://example.com"),
	}.Build()

	apiKey := configv1.APIKeyAuth_builder{
		ParamName:         proto.String("X-API-Key"),
		VerificationValue: proto.String("some-value"),
	}.Build()

	authConfig := configv1.Authentication_builder{
		ApiKey: apiKey,
	}.Build()

	svcConfig := configv1.UpstreamServiceConfig_builder{
		Name:         proto.String("test-service"),
		HttpService:  httpSvc,
		UpstreamAuth: authConfig,
	}.Build()

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{svcConfig},
	}.Build()

	errs := Validate(ctx, config, Server)

	if len(errs) == 0 {
		t.Fatal("Expected validation error for upstream API key missing 'value', but got none")
	}

	assert.Contains(t, errs[0].Err.Error(), "api key 'value' is missing")
}
