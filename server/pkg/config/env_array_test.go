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

func TestEnvVarOverrideArray(t *testing.T) {
	// Create a temporary config file
	fs := afero.NewMemMapFs()
	configContent := `
global_settings:
  profiles:
    - default
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	_ = os.Setenv("MCPANY__GLOBAL_SETTINGS__PROFILES", "admin,user")
	defer func() {
		_ = os.Unsetenv("MCPANY__GLOBAL_SETTINGS__PROFILES")
	}()

	// Load the config
	store := config.NewFileStore(fs, []string{"/config.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	// Verify overrides
	require.NotNil(t, cfg.GetGlobalSettings().GetProfiles())
	assert.Contains(t, cfg.GetGlobalSettings().GetProfiles(), "admin")
	assert.Contains(t, cfg.GetGlobalSettings().GetProfiles(), "user")
    assert.Len(t, cfg.GetGlobalSettings().GetProfiles(), 2)
}
