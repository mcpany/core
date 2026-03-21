// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvVarListOverride(t *testing.T) {
	// Create a temporary config file
	fs := afero.NewMemMapFs()
	configContent := `
global_settings:
  profiles:
    - default
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables to override the list
	// We want to replace the list with ["custom1", "custom2"]
	_ = os.Setenv("MCPANY__GLOBAL_SETTINGS__PROFILES__0", "custom1")
	_ = os.Setenv("MCPANY__GLOBAL_SETTINGS__PROFILES__1", "custom2")
	defer func() {
		_ = os.Unsetenv("MCPANY__GLOBAL_SETTINGS__PROFILES__0")
		_ = os.Unsetenv("MCPANY__GLOBAL_SETTINGS__PROFILES__1")
	}()

	// Load the config
	store := config.NewFileStore(fs, []string{"/config.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	// Verify overrides
	require.NotNil(t, cfg.GetGlobalSettings().GetProfiles())
	assert.Equal(t, []string{"custom1", "custom2"}, cfg.GetGlobalSettings().GetProfiles())
}
