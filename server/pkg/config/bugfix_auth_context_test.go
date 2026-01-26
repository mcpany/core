// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestValidate_UpstreamApiKeyMissingValue(t *testing.T) {
	ctx := context.Background()

	// Upstream Service with API Key auth that has verification_value but NO value.
	// This should be invalid for an UPSTREAM service (outgoing).
	httpSvc := &configv1.HttpUpstreamService{}
	httpSvc.SetAddress("http://example.com")

	apiKey := &configv1.APIKeyAuth{}
	apiKey.SetParamName("X-API-Key")
	apiKey.SetVerificationValue("some-value")

	authConfig := &configv1.Authentication{}
	authConfig.SetApiKey(apiKey)

	svcConfig := &configv1.UpstreamServiceConfig{}
	svcConfig.SetName("test-service")
	svcConfig.SetHttpService(httpSvc)
	svcConfig.SetUpstreamAuth(authConfig)

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{svcConfig},
	}.Build()

	errs := Validate(ctx, config, Server)

	if len(errs) == 0 {
		t.Fatal("Expected validation error for upstream API key missing 'value', but got none")
	}

	assert.Contains(t, errs[0].Err.Error(), "api key 'value' is missing")
}
