// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/app"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDrift(t *testing.T) {
	// 1. Setup initial config
	fs := afero.NewMemMapFs()
	configContent := `
global_settings:
  mcp_listen_address: "localhost:0" # Let OS pick port
upstream_services:
  - name: "service-a"
    http_service:
      address: "http://example.com"
      tools:
        - name: "tool-a"
          call_id: "call-a"
      calls:
        call-a:
          method: HTTP_METHOD_GET
          endpoint_path: "/a"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0600)
	require.NoError(t, err)

	appInstance := app.NewApplication()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Start Application in background
	go func() {
		// Use empty ports to avoid conflicts, and mock filesystem
		err := appInstance.Run(ctx, fs, false, ":0", "", []string{"/config.yaml"}, 1*time.Second)
		if err != nil && ctx.Err() == nil {
			t.Logf("Run exited with error: %v", err) // Use Logf to avoid race with t.Error if test ends
		}
	}()

	// Wait for startup (ServiceRegistry populated)
	require.Eventually(t, func() bool {
		if appInstance.ServiceRegistry == nil {
			return false
		}
		services, err := appInstance.ServiceRegistry.GetAllServices()
		if err != nil {
			return false
		}
		for _, s := range services {
			if s.GetName() == "service-a" {
				return true
			}
		}
		return false
	}, 5*time.Second, 100*time.Millisecond)

	// 3. Verify Initial State
	services, err := appInstance.ServiceRegistry.GetAllServices()
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "service-a", services[0].GetName())

	// 4. Update Config (Remove service-a)
	newConfigContent := `
global_settings:
  mcp_listen_address: "localhost:0"
upstream_services: []
`
	err = afero.WriteFile(fs, "/config.yaml", []byte(newConfigContent), 0600)
	require.NoError(t, err)

	// 5. Reload Config
	err = appInstance.ReloadConfig(fs, []string{"/config.yaml"})
	require.NoError(t, err)

	// 6. Verify Drift (Bug Reproduction)
	// ToolManager should be empty (ReloadConfig clears it)
	tools := appInstance.ToolManager.ListTools()
	assert.Empty(t, tools, "ToolManager should be empty")

	// ServiceRegistry should ALSO be empty (This is what we expect to FAIL currently)
	services, err = appInstance.ServiceRegistry.GetAllServices()
	require.NoError(t, err)

	if len(services) > 0 {
		assert.Fail(t, "Bug Reproduced: ServiceRegistry still has services after removal in config")
	}
}
