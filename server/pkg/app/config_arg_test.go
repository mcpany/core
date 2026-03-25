// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigLoadingWithArg(t *testing.T) {
	// Create a temporary config file
	configContent := `
upstream_services:
  - name: test-config-arg
    http_service:
      address: http://localhost:8080
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Ensure env var is NOT set, and restore it later to avoid side effects
	oldEnv, hasEnv := os.LookupEnv("MCPANY_ENABLE_FILE_CONFIG")
	if hasEnv {
		defer os.Setenv("MCPANY_ENABLE_FILE_CONFIG", oldEnv)
	} else {
		defer os.Unsetenv("MCPANY_ENABLE_FILE_CONFIG")
	}
	os.Unsetenv("MCPANY_ENABLE_FILE_CONFIG")

	app := NewApplication()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run app in goroutine
	go func() {
		_ = app.Run(RunOptions{
			Ctx:         ctx,
			Fs:          afero.NewOsFs(),
			ConfigPaths: []string{tmpFile.Name()},
			// Use random ports to avoid conflicts
			JSONRPCPort: "localhost:0",
			GRPCPort:    "localhost:0",
		})
	}()

	// Wait for startup
	select {
	case <-app.startupCh:
		// Startup complete
	case <-time.After(5 * time.Second):
		t.Fatal("Startup timed out")
	}

	// Check if service was loaded
	require.NotNil(t, app.ServiceRegistry, "ServiceRegistry should be initialized")
	services, err := app.ServiceRegistry.GetAllServices()
	require.NoError(t, err)

	found := false
	for _, svc := range services {
		if svc.GetName() == "test-config-arg" {
			found = true
			break
		}
	}

	assert.True(t, found, "Service 'test-config-arg' should be loaded when --config-path is provided")
}
