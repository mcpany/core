// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestValidate_McpBundleConnection_Security(t *testing.T) {
	t.Run("invalid bundle connection - insecure path with ..", func(t *testing.T) {
		cfg := func() *configv1.McpAnyServerConfig {
			svc := configv1.UpstreamServiceConfig_builder{
				Name: proto.String("mcp-bundle-svc-insecure-dotdot"),
				McpService: configv1.McpUpstreamService_builder{
					BundleConnection: configv1.McpBundleConnection_builder{
						BundlePath: proto.String("../insecure.bundle"),
					}.Build(),
				}.Build(),
			}.Build()

			return configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
			}.Build()
		}()

		validationErrors := Validate(context.Background(), cfg, Server)
		assert.NotEmpty(t, validationErrors, "Should return validation error for insecure path")
		if len(validationErrors) > 0 {
			assert.Contains(t, validationErrors[0].Error(), "insecure bundle_path")
		}
	})

	t.Run("invalid bundle connection - absolute path not in allowed list", func(t *testing.T) {
		cfg := func() *configv1.McpAnyServerConfig {
			svc := configv1.UpstreamServiceConfig_builder{
				Name: proto.String("mcp-bundle-svc-absolute"),
				McpService: configv1.McpUpstreamService_builder{
					BundleConnection: configv1.McpBundleConnection_builder{
						BundlePath: proto.String("/etc/passwd"),
					}.Build(),
				}.Build(),
			}.Build()

			return configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
			}.Build()
		}()

		validationErrors := Validate(context.Background(), cfg, Server)
		assert.NotEmpty(t, validationErrors, "Should return validation error for absolute path")
		if len(validationErrors) > 0 {
			assert.Contains(t, validationErrors[0].Error(), "insecure bundle_path")
		}
	})
}
