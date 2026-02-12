// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDocCmd(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
global_settings:
  mcp_listen_address: ":50050"
  log_level: "info"

upstream_services:
  - name: "my-filesystem"
    filesystem_service:
      root_paths:
        "/data": "/tmp/data"
      read_only: true
  - name: "my-http"
    http_service:
      address: "http://example.com"
`
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	// Setup command
	cmd := newConfigDocCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)

	// Bind flags so config.Load works
	config.BindRootFlags(cmd)

	cmd.SetArgs([]string{"--config-path", configFile})

	// Execute
	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()

	// Assertions
	assert.Contains(t, output, "# MCP Server Configuration")
	assert.Contains(t, output, "Listen Address")
	assert.Contains(t, output, ":50050")
	assert.Contains(t, output, "Upstream Services")

	// Check Filesystem
	assert.Contains(t, output, "### my-filesystem")
	assert.Contains(t, output, "**Type**: Filesystem")
	assert.Contains(t, output, "`list_directory`")
	assert.Contains(t, output, "`read_file`")
	// read_only is true, so write_file should NOT be present
	assert.NotContains(t, output, "`write_file`")

	// Check HTTP
	assert.Contains(t, output, "### my-http")
	assert.Contains(t, output, "**Type**: HTTP")
	assert.Contains(t, output, "http://example.com")
}
