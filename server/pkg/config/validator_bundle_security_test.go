package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestValidate_McpBundleConnection_Security(t *testing.T) {
	t.Run("invalid bundle connection - insecure path with ..", func(t *testing.T) {
		cfg := func() *configv1.McpAnyServerConfig {
			cfg := &configv1.McpAnyServerConfig{}
			svc := &configv1.UpstreamServiceConfig{}
			svc.SetName("mcp-bundle-svc-insecure-dotdot")

			mcp := &configv1.McpUpstreamService{}
			conn := &configv1.McpBundleConnection{}
			conn.SetBundlePath("../insecure.bundle")
			mcp.SetBundleConnection(conn)
			svc.SetMcpService(mcp)

			cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{svc})
			return cfg
		}()

		validationErrors := Validate(context.Background(), cfg, Server)
		assert.NotEmpty(t, validationErrors, "Should return validation error for insecure path")
		if len(validationErrors) > 0 {
			assert.Contains(t, validationErrors[0].Error(), "insecure bundle_path")
		}
	})

	t.Run("invalid bundle connection - absolute path not in allowed list", func(t *testing.T) {
		cfg := func() *configv1.McpAnyServerConfig {
			cfg := &configv1.McpAnyServerConfig{}
			svc := &configv1.UpstreamServiceConfig{}
			svc.SetName("mcp-bundle-svc-absolute")

			mcp := &configv1.McpUpstreamService{}
			conn := &configv1.McpBundleConnection{}
			conn.SetBundlePath("/etc/passwd")
			mcp.SetBundleConnection(conn)
			svc.SetMcpService(mcp)

			cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{svc})
			return cfg
		}()

		validationErrors := Validate(context.Background(), cfg, Server)
		assert.NotEmpty(t, validationErrors, "Should return validation error for absolute path")
		if len(validationErrors) > 0 {
			assert.Contains(t, validationErrors[0].Error(), "insecure bundle_path")
		}
	})
}
