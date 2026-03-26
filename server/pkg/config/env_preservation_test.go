// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvVarOverridePreservesOtherListItems(t *testing.T) {
	// Create a temporary config file with two services
	fs := afero.NewMemMapFs()
	configContent := `
upstream_services:
  - name: "service-1"
    http_service:
      address: "http://service1.com"
  - name: "service-2"
    http_service:
      address: "http://service2.com"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// Override URL of service-1
	envVar := "MCPANY__UPSTREAM_SERVICES__0__HTTP_SERVICE__ADDRESS"
	val := "http://service1-overridden.com"

	_ = os.Setenv(envVar, val)
	defer func() {
		_ = os.Unsetenv(envVar)
	}()

	// Load the config
	store := config.NewFileStore(fs, []string{"/config.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	// Verify we still have 2 services
	require.Len(t, cfg.GetUpstreamServices(), 2, "Should have 2 services")

	// Verify service 1 is overridden
	assert.Equal(t, "service-1", cfg.GetUpstreamServices()[0].GetName(), "Service 1 name should be preserved")
	assert.Equal(t, "http://service1-overridden.com", cfg.GetUpstreamServices()[0].GetHttpService().GetAddress())

	// Verify service 2 is preserved
	assert.Equal(t, "service-2", cfg.GetUpstreamServices()[1].GetName())
	assert.Equal(t, "http://service2.com", cfg.GetUpstreamServices()[1].GetHttpService().GetAddress())
}
