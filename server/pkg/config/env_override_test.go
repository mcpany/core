// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvVarOverride(t *testing.T) {
	// Create a temporary config file
	fs := afero.NewMemMapFs()
	configContent := `
global_settings:
  log_level: INFO
  mcp_listen_address: "127.0.0.1:50050"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables to override settings
	_ = os.Setenv("MCPANY__GLOBAL_SETTINGS__LOG_LEVEL", "DEBUG")
	_ = os.Setenv("MCPANY__GLOBAL_SETTINGS__MCP_LISTEN_ADDRESS", "0.0.0.0:6000")
	defer func() {
		_ = os.Unsetenv("MCPANY__GLOBAL_SETTINGS__LOG_LEVEL")
		_ = os.Unsetenv("MCPANY__GLOBAL_SETTINGS__MCP_LISTEN_ADDRESS")
	}()

	// Load the config
	store := config.NewFileStore(fs, []string{"/config.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	// Verify overrides
	require.NotNil(t, cfg.GetGlobalSettings().GetLogLevel())
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_DEBUG, cfg.GetGlobalSettings().GetLogLevel())
	require.NotNil(t, cfg.GetGlobalSettings().GetMcpListenAddress())
	assert.Equal(t, "0.0.0.0:6000", cfg.GetGlobalSettings().GetMcpListenAddress())
}

func TestEnvVarInjectionWithoutConfig(t *testing.T) {
	// Create an empty config file
	fs := afero.NewMemMapFs()
	// We need a valid YAML file that results in a map, typically {}
	err := afero.WriteFile(fs, "/minimal.yaml", []byte("{}"), 0644)
	require.NoError(t, err)

	// Set environment variables
	_ = os.Setenv("MCPANY__GLOBAL_SETTINGS__MCP_LISTEN_ADDRESS", "0.0.0.0:7000")
	defer func() {
		_ = os.Unsetenv("MCPANY__GLOBAL_SETTINGS__MCP_LISTEN_ADDRESS")
	}()

	// Load the config
	store := config.NewFileStore(fs, []string{"/minimal.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	require.NotNil(t, cfg.GetGlobalSettings().GetMcpListenAddress())
	assert.Equal(t, "0.0.0.0:7000", cfg.GetGlobalSettings().GetMcpListenAddress())
}

func TestEnvVarOverrideLowercase(t *testing.T) {
	// Create a temporary config file
	fs := afero.NewMemMapFs()
	configContent := `
global_settings:
  log_level: INFO
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables to override settings with lowercase
	_ = os.Setenv("MCPANY__GLOBAL_SETTINGS__LOG_LEVEL", "debug")
	defer func() {
		_ = os.Unsetenv("MCPANY__GLOBAL_SETTINGS__LOG_LEVEL")
	}()

	// Load the config
	store := config.NewFileStore(fs, []string{"/config.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	// Verify overrides
	require.NotNil(t, cfg.GetGlobalSettings().GetLogLevel())
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_DEBUG, cfg.GetGlobalSettings().GetLogLevel())
}
