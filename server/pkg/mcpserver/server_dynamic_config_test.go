// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/app"
	"github.com/mcpany/core/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamicConfigReload(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Initial Config
	initialConfig := `
global_settings:
  mcp_listen_address: ":0" # Random port
  log_level: LOG_LEVEL_DEBUG
upstream_services:
  - name: "service-a"
    id: "service-a-id"
    command_line_service:
      command: "echo"
      calls:
        call-a:
          args: ["hello"]
      tools:
        - name: "tool-a"
          description: "Tool A"
          call_id: "call-a"
`
	err := os.WriteFile(configFile, []byte(initialConfig), 0644)
	require.NoError(t, err)

	// Setup Application
	application := app.NewApplication()
	fs := afero.NewOsFs()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run application in a goroutine
	go func() {
		_ = application.Run(ctx, fs, false, ":0", "", []string{configFile}, 1*time.Second)
	}()

	// Wait for server to start
	// In a real test we'd wait for a ready signal, but sleep is simple for now
	time.Sleep(2 * time.Second)

	// Verify Tool A exists
	tools := application.ToolManager.ListTools()
	foundA := false
	for _, t := range tools {
		if t.Tool().GetName() == "tool-a" {
			foundA = true
			break
		}
	}
	assert.True(t, foundA, "Tool A should be present initially")

	// Setup Watcher manually since main.go logic isn't running here.
	watcher, err := config.NewWatcher()
	require.NoError(t, err)
	defer watcher.Close()

	err = watcher.Add(configFile)
	require.NoError(t, err)

	watcher.Start(ctx, 100*time.Millisecond, func() error {
		return application.ReloadConfig(fs, []string{configFile})
	})

	// Modify Config to add Tool B
	newConfig := `
global_settings:
  mcp_listen_address: ":0"
  log_level: LOG_LEVEL_DEBUG
upstream_services:
  - name: "service-a"
    id: "service-a-id"
    command_line_service:
      command: "echo"
      calls:
        call-a:
          args: ["hello"]
      tools:
        - name: "tool-a"
          description: "Tool A"
          call_id: "call-a"
  - name: "service-b"
    id: "service-b-id"
    command_line_service:
      command: "echo"
      calls:
        call-b:
          args: ["hello"]
      tools:
        - name: "tool-b"
          description: "Tool B"
          call_id: "call-b"
`
	err = os.WriteFile(configFile, []byte(newConfig), 0644)
	require.NoError(t, err)

	// Wait for reload
	time.Sleep(1 * time.Second)

	// Verify Tool B exists
	tools = application.ToolManager.ListTools()
	foundB := false
	foundA = false
	for _, t := range tools {
		if t.Tool().GetName() == "tool-b" {
			foundB = true
		}
		if t.Tool().GetName() == "tool-a" {
			foundA = true
		}
	}
	assert.True(t, foundA, "Tool A should still be present")
	assert.True(t, foundB, "Tool B should be present after reload")
}
