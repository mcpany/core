// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestReloadConfig_Directory(t *testing.T) {
	fs := afero.NewMemMapFs()
	app := NewApplication()

	// Mock ServiceRegistry
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	app.ServiceRegistry = serviceregistry.New(
		upstreamFactory,
		app.ToolManager,
		app.PromptManager,
		app.ResourceManager,
		auth.NewManager(),
	)

	// Create a directory with config files
	err := fs.MkdirAll("/config", 0o755)
	require.NoError(t, err)

	configContent1 := `
upstream_services:
 - name: "service1"
   http_service:
     address: "http://localhost:8081"
`
	err = afero.WriteFile(fs, "/config/service1.yaml", []byte(configContent1), 0o644)
	require.NoError(t, err)

	// We use yaml content but with .json extension?
	// The loader might expect JSON format for .json.
	// Let's use valid JSON for the json file.
	configContent2 := `
{
  "upstream_services": [
    {
      "name": "service2",
      "http_service": {
        "address": "http://localhost:8082"
      }
    }
  ]
}
`
	err = afero.WriteFile(fs, "/config/service2.json", []byte(configContent2), 0o644)
	require.NoError(t, err)

	// Write a file that should be ignored by readConfigFiles walker
	err = afero.WriteFile(fs, "/config/README.md", []byte("# Config"), 0o644)
	require.NoError(t, err)

	err = app.ReloadConfig(context.Background(), fs, []string{"/config"})
	require.NoError(t, err)
}
