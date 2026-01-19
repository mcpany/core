// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRun_FileConfigEnabledByPath(t *testing.T) {
	// This test verifies that passing a config path enables file config loading
	// even without the environment variable.
	t.Setenv("MCPANY_ENABLE_FILE_CONFIG", "false")

	fs := afero.NewMemMapFs()
	// Create a config file that defines a service
	configContent := `
upstream_services:
  - name: "enabled-by-path"
    http_service:
      address: "http://localhost:8080"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app := NewApplication()
	// Need a real store or mock that doesn't interfere
	mockStore := new(MockStore)
	mockStore.On("Load", mock.Anything).Return((*configv1.McpAnyServerConfig)(nil), nil)
	mockStore.On("ListServices", mock.Anything).Return([]*configv1.UpstreamServiceConfig{}, nil)
	mockStore.On("GetGlobalSettings", mock.Anything).Return(&configv1.GlobalSettings{}, nil)
	mockStore.On("Close").Return(nil)
	app.Storage = mockStore

	// Run in a goroutine because it blocks
	go func() {
		// We use an empty API key to allow public access logic if any, but test focuses on config loading.
		_ = app.Run(ctx, fs, false, "localhost:0", "localhost:0", []string{"/config.yaml"}, "", 5*time.Second)
	}()

	// Wait for startup
	err = app.WaitForStartup(ctx)
	require.NoError(t, err)

	// Check if service is registered
	// app.ServiceRegistry is set now
	services, err := app.ServiceRegistry.GetAllServices()
	require.NoError(t, err)

	found := false
	for _, s := range services {
		if s.GetName() == "enabled-by-path" {
			found = true
			break
		}
	}
	assert.True(t, found, "Service from config file should be loaded even if MCPANY_ENABLE_FILE_CONFIG is false, because config path was explicitly provided")
}
