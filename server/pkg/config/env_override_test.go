// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	configv1 "github.com/mcpany/core/proto/config/v1"
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
	require.NotNil(t, cfg.GlobalSettings.LogLevel)
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_DEBUG, *cfg.GlobalSettings.LogLevel)
	require.NotNil(t, cfg.GlobalSettings.McpListenAddress)
	assert.Equal(t, "0.0.0.0:6000", *cfg.GlobalSettings.McpListenAddress)
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

	require.NotNil(t, cfg.GlobalSettings.McpListenAddress)
	assert.Equal(t, "0.0.0.0:7000", *cfg.GlobalSettings.McpListenAddress)
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
	require.NotNil(t, cfg.GlobalSettings.LogLevel)
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_DEBUG, *cfg.GlobalSettings.LogLevel)
}

func TestEnvVarRepeatedMessage(t *testing.T) {
	// Create an empty config file
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/minimal.yaml", []byte("{}"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		envValue string
		verify   func(t *testing.T, cfg *configv1.McpAnyServerConfig)
	}{
		{
			name:     "JSON Array",
			envValue: `[{"name": "service1", "http_service": {"address": "http://example.com"}}]`,
			verify: func(t *testing.T, cfg *configv1.McpAnyServerConfig) {
				require.Len(t, cfg.UpstreamServices, 1)
				assert.Equal(t, "service1", *cfg.UpstreamServices[0].Name)
			},
		},
		{
			name:     "JSON Object (Single)",
			envValue: `{"name": "service2", "http_service": {"address": "http://example.org"}}`,
			verify: func(t *testing.T, cfg *configv1.McpAnyServerConfig) {
				require.Len(t, cfg.UpstreamServices, 1)
				assert.Equal(t, "service2", *cfg.UpstreamServices[0].Name)
			},
		},
		{
			name:     "Malformed JSON Array (Fallback to CSV)",
			envValue: `[{"name": "service3"`, // Invalid JSON
			verify: func(t *testing.T, cfg *configv1.McpAnyServerConfig) {
                // If the fallback kicks in, it treats the whole string as one item (or split by comma).
                // `[{"name": "service3"` is not a valid JSON for the object.
                // So it will be appended as string.
                // protojson unmarshal will likely fail or ignore it.
                // In any case, we want to ensure it doesn't crash.
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("MCPANY__UPSTREAM_SERVICES", tt.envValue)
			defer os.Unsetenv("MCPANY__UPSTREAM_SERVICES")

			store := config.NewFileStore(fs, []string{"/minimal.yaml"})
			cfg, err := store.Load(context.Background())

            if tt.name == "Malformed JSON Array (Fallback to CSV)" {
                assert.Error(t, err)
            } else {
                require.NoError(t, err)
                tt.verify(t, cfg)
            }
		})
	}
}

func TestEnvVarRepeatedMessageCSVWithJSON(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/minimal.yaml", []byte("{}"), 0644)
	require.NoError(t, err)

	os.Setenv("MCPANY__UPSTREAM_SERVICES", `{"name": "s1"}, {"name": "s2"}`)
	defer os.Unsetenv("MCPANY__UPSTREAM_SERVICES")

	store := config.NewFileStore(fs, []string{"/minimal.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

    require.Len(t, cfg.UpstreamServices, 2)
    assert.Equal(t, "s1", *cfg.UpstreamServices[0].Name)
    assert.Equal(t, "s2", *cfg.UpstreamServices[1].Name)
}
