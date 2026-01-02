// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"os"
	"testing"

	"github.com/mcpany/core/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvVarOverrideList(t *testing.T) {
	// Create a temporary config file
	fs := afero.NewMemMapFs()
	configContent := `
global_settings:
  log_level: INFO
  mcp_listen_address: "127.0.0.1:50050"
upstream_services:
  - name: "service1"
    http_service:
      address: "http://example.com"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables to override settings
	_ = os.Setenv("MCPANY__UPSTREAM_SERVICES__0__HTTP_SERVICE__ADDRESS", "http://overridden.com")
	defer func() {
		_ = os.Unsetenv("MCPANY__UPSTREAM_SERVICES__0__HTTP_SERVICE__ADDRESS")
	}()

	// Load the config
	store := config.NewFileStore(fs, []string{"/config.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	// Check if override happened
	require.Len(t, cfg.UpstreamServices, 1)

	httpSvc := cfg.UpstreamServices[0].GetHttpService()
	require.NotNil(t, httpSvc)
	// Use GetAddress() accessor which safely returns string value
	assert.Equal(t, "http://overridden.com", httpSvc.GetAddress())
}
