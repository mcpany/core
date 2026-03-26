// Copyright 2026 Author(s) of MCP Any
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

func TestEnvVarOverrideBoolean(t *testing.T) {
	// Create a temporary config file
	fs := afero.NewMemMapFs()
	configContent := `
global_settings:
  audit:
    enabled: false
    log_arguments: false
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables to override settings
	_ = os.Setenv("MCPANY__GLOBAL_SETTINGS__AUDIT__ENABLED", "true")
	_ = os.Setenv("MCPANY__GLOBAL_SETTINGS__AUDIT__LOG_ARGUMENTS", "true")
	defer func() {
		_ = os.Unsetenv("MCPANY__GLOBAL_SETTINGS__AUDIT__ENABLED")
		_ = os.Unsetenv("MCPANY__GLOBAL_SETTINGS__AUDIT__LOG_ARGUMENTS")
	}()

	// Load the config
	store := config.NewFileStore(fs, []string{"/config.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	// Verify overrides
	require.NotNil(t, cfg.GetGlobalSettings().GetAudit())
	assert.True(t, cfg.GetGlobalSettings().GetAudit().GetEnabled(), "Audit should be enabled")
	assert.True(t, cfg.GetGlobalSettings().GetAudit().GetLogArguments(), "LogArguments should be enabled")
}
