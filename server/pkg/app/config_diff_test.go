// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDiffGeneration(t *testing.T) {
	app := NewApplication()
	fs := afero.NewMemMapFs()
	ctx := context.Background()
	configPath := "config.yaml"

	// 1. Initial valid config
	initialConfig := `
upstream_services:
  - name: "echo"
    http_service:
      address: "http://echo.service"
`
	err := afero.WriteFile(fs, configPath, []byte(initialConfig), 0644)
	require.NoError(t, err)

	// Reload (Initial load)
	err = app.ReloadConfig(ctx, fs, []string{configPath})
	require.NoError(t, err)
	assert.NotEmpty(t, app.lastGoodConfig)
	assert.Empty(t, app.configDiff)

	// 2. Modify to invalid config (syntax error)
	invalidConfig := `
upstream_services:
  - name: "echo"
    http_service:
      address: "http://echo.service"
  - invalid_indentation
`
	err = afero.WriteFile(fs, configPath, []byte(invalidConfig), 0644)
	require.NoError(t, err)

	// Reload (Should fail)
	err = app.ReloadConfig(ctx, fs, []string{configPath})
	assert.Error(t, err)
	assert.NotEmpty(t, app.configDiff)
	assert.Contains(t, app.configDiff, "invalid_indentation")
	// Verify it shows diff
	assert.Contains(t, app.configDiff, "+  - invalid_indentation")

	// 3. Fix config (Valid again)
	// Changing echo service to verify we track changes
	fixedConfig := `
upstream_services:
  - name: "echo-v2"
    http_service:
      address: "http://echo.service/v2"
`
	err = afero.WriteFile(fs, configPath, []byte(fixedConfig), 0644)
	require.NoError(t, err)

	// Reload (Should succeed)
	err = app.ReloadConfig(ctx, fs, []string{configPath})
	require.NoError(t, err)
	assert.Empty(t, app.configDiff)
	assert.Equal(t, fixedConfig, app.lastGoodConfig[configPath])
}

func TestConfigDiffNewFile(t *testing.T) {
	app := NewApplication()
	fs := afero.NewMemMapFs()
	ctx := context.Background()
	configPath1 := "config1.yaml"
	configPath2 := "config2.yaml"

	// 1. Initial config
	err := afero.WriteFile(fs, configPath1, []byte("upstream_services: []"), 0644)
	require.NoError(t, err)

	err = app.ReloadConfig(ctx, fs, []string{configPath1})
	require.NoError(t, err)

	// 2. Add new invalid file
	err = afero.WriteFile(fs, configPath2, []byte("invalid_yaml"), 0644)
	require.NoError(t, err)

	// Reload with both files
	err = app.ReloadConfig(ctx, fs, []string{configPath1, configPath2})
	assert.Error(t, err)
	assert.Contains(t, app.configDiff, "New file: config2.yaml")
	assert.Contains(t, app.configDiff, "+invalid_yaml")
}
