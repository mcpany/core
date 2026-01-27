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
		cfg := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("mcp-bundle-svc-insecure-dotdot"),
					ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
						McpService: &configv1.McpUpstreamService{
							ConnectionType: &configv1.McpUpstreamService_BundleConnection{
								BundleConnection: &configv1.McpBundleConnection{
									BundlePath: proto.String("../insecure.bundle"),
								},
							},
						},
					},
				},
			},
		}

		validationErrors := Validate(context.Background(), cfg, Server)
		assert.NotEmpty(t, validationErrors, "Should return validation error for insecure path")
		if len(validationErrors) > 0 {
			assert.Contains(t, validationErrors[0].Error(), "insecure bundle_path")
		}
	})

	t.Run("invalid bundle connection - absolute path not in allowed list", func(t *testing.T) {
		cfg := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("mcp-bundle-svc-absolute"),
					ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
						McpService: &configv1.McpUpstreamService{
							ConnectionType: &configv1.McpUpstreamService_BundleConnection{
								BundleConnection: &configv1.McpBundleConnection{
									BundlePath: proto.String("/etc/passwd"),
								},
							},
						},
					},
				},
			},
		}

		validationErrors := Validate(context.Background(), cfg, Server)
		assert.NotEmpty(t, validationErrors, "Should return validation error for absolute path")
		if len(validationErrors) > 0 {
			assert.Contains(t, validationErrors[0].Error(), "insecure bundle_path")
		}
	})
}
