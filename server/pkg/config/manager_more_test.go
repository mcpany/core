// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamServiceManager_ApplyAuthentication_Error(t *testing.T) {
	mgr := &UpstreamServiceManager{}
	req := httptest.NewRequest("GET", "http://example.com", nil)

	// Auth with non-existent env var
	auth := configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			Value: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("NON_EXISTENT_VAR_FOR_TESTING"),
			}.Build(),
			ParamName: proto.String("X-API-Key"),
		}.Build(),
	}.Build()

	err := mgr.applyAuthentication(req, auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment variable \"NON_EXISTENT_VAR_FOR_TESTING\" is not set")
}
