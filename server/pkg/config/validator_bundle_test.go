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

func TestValidate_McpBundleConnection(t *testing.T) {
	t.Run("valid bundle connection", func(t *testing.T) {
		cfg := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("mcp-bundle-svc"),
					ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
						McpService: &configv1.McpUpstreamService{
							ToolAutoDiscovery: proto.Bool(true),
							ConnectionType: &configv1.McpUpstreamService_BundleConnection{
								BundleConnection: &configv1.McpBundleConnection{
									BundlePath: proto.String("test.bundle"),
								},
							},
						},
					},
				},
			},
		}

		validationErrors := Validate(context.Background(), cfg, Server)
		assert.Empty(t, validationErrors)
	})

	t.Run("invalid bundle connection - empty path", func(t *testing.T) {
		cfg := &configv1.McpAnyServerConfig{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				{
					Name: proto.String("mcp-bundle-svc-invalid"),
					ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
						McpService: &configv1.McpUpstreamService{
							ConnectionType: &configv1.McpUpstreamService_BundleConnection{
								BundleConnection: &configv1.McpBundleConnection{
									BundlePath: proto.String(""),
								},
							},
						},
					},
				},
			},
		}

		validationErrors := Validate(context.Background(), cfg, Server)
		assert.NotEmpty(t, validationErrors)
		assert.Contains(t, validationErrors[0].Error(), "mcp service with bundle_connection has empty bundle_path")
	})
}
