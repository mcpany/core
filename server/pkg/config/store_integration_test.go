// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_RelativePathResolution(t *testing.T) {
	fs := afero.NewMemMapFs()
	baseDir := "/config/subdir"
	require.NoError(t, fs.MkdirAll(baseDir, 0755))

	configFile := filepath.Join(baseDir, "mcp.yaml")
	configContent := `
upstream_services:
  - name: script-service
    command_line_service:
      command: "./script.sh"
      working_directory: "data"
`
	require.NoError(t, afero.WriteFile(fs, configFile, []byte(configContent), 0644))

	store := NewFileStore(fs, []string{configFile})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	require.Len(t, cfg.UpstreamServices, 1)
	svc := cfg.UpstreamServices[0].GetCommandLineService()
	require.NotNil(t, svc)

	// Verify paths are absolute and correct
	assert.Equal(t, filepath.Join(baseDir, "./script.sh"), svc.GetCommand())
	assert.Equal(t, filepath.Join(baseDir, "data"), svc.GetWorkingDirectory())
}

func TestFileStore_RelativePathResolution_NoSeparator(t *testing.T) {
	fs := afero.NewMemMapFs()
	baseDir := "/config"
	require.NoError(t, fs.MkdirAll(baseDir, 0755))

	configFile := filepath.Join(baseDir, "mcp.yaml")
	// "npm" has no separator, so it should NOT be resolved
	configContent := `
upstream_services:
  - name: npm-service
    command_line_service:
      command: "npm"
`
	require.NoError(t, afero.WriteFile(fs, configFile, []byte(configContent), 0644))

	store := NewFileStore(fs, []string{configFile})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	require.Len(t, cfg.UpstreamServices, 1)
	svc := cfg.UpstreamServices[0].GetCommandLineService()
	require.NotNil(t, svc)

	// Verify "npm" is unchanged
	assert.Equal(t, "npm", svc.GetCommand())
}
