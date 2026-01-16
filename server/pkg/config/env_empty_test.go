package config_test

import (
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/spf13/afero"
)

// TestProfilesEmptyEnv verifies that setting the MCPANY_PROFILES environment variable
// to an empty string correctly clears the profiles list, instead of falling back to default.
func TestProfilesEmptyEnv(t *testing.T) {
	// Reset viper
	viper.Reset()
	fs := afero.NewMemMapFs()

	cmd := &cobra.Command{}
	config.BindFlags(cmd)

	// Set env var to empty string to try to clear profiles
	os.Setenv("MCPANY_PROFILES", "")
	defer os.Unsetenv("MCPANY_PROFILES")

	settings := config.GlobalSettings()
	err := settings.Load(cmd, fs)
	require.NoError(t, err)

	profiles := settings.Profiles()
	assert.Empty(t, profiles, "Profiles should be empty when env var is empty string")
}
