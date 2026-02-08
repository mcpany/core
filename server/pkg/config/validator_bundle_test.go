package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	proto "google.golang.org/protobuf/proto"
)

func TestValidate_McpBundleConnection(t *testing.T) {
	t.Run("valid bundle connection", func(t *testing.T) {
		cfg := configv1.McpAnyServerConfig_builder{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				configv1.UpstreamServiceConfig_builder{
					Name: proto.String("mcp-bundle-svc"),
					McpService: configv1.McpUpstreamService_builder{
						BundleConnection: configv1.McpBundleConnection_builder{
							BundlePath: proto.String("test.bundle"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		validationErrors := Validate(context.Background(), cfg, Server)
		assert.Empty(t, validationErrors)
	})

	t.Run("invalid bundle connection - empty path", func(t *testing.T) {
		cfg := configv1.McpAnyServerConfig_builder{
			UpstreamServices: []*configv1.UpstreamServiceConfig{
				configv1.UpstreamServiceConfig_builder{
					Name: proto.String("mcp-bundle-svc-invalid"),
					McpService: configv1.McpUpstreamService_builder{
						BundleConnection: configv1.McpBundleConnection_builder{
							BundlePath: proto.String(""),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		validationErrors := Validate(context.Background(), cfg, Server)
		assert.NotEmpty(t, validationErrors)
		assert.Contains(t, validationErrors[0].Error(), "mcp service with bundle_connection has empty bundle_path")
	})
}
