package config

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestProfileOverride(t *testing.T) {
	// Reset viper
	viper.Reset()
	// Bind flags
	cmd := &cobra.Command{}
	BindFlags(cmd)

	// Set Env Var (simulating flag)
	os.Setenv("MCPANY_PROFILES", "env-profile")
	defer os.Unsetenv("MCPANY_PROFILES")

	// Create Config File with different profile
	fs := afero.NewMemMapFs()
	configContent := `
global_settings:
  profiles: ["file-profile"]
upstream_services: []
`
	_ = afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)

	// Load GlobalSettings
	gs := GlobalSettings()
	err := gs.Load(cmd, fs)
	assert.NoError(t, err)

	assert.Equal(t, []string{"env-profile"}, gs.Profiles())

    // Now LoadResolvedConfig
    store := NewFileStore(fs, []string{"config.yaml"})

    // We expect LoadResolvedConfig to prioritize GlobalSettings (Env) over FileConfig
    cfg, err := LoadResolvedConfig(context.Background(), store)
    assert.NoError(t, err)

    // Assert the CORRECT behavior: env wins
    expectedProfiles := []string{"env-profile"}
    actualProfiles := cfg.GetGlobalSettings().GetProfiles()

    assert.Equal(t, expectedProfiles, actualProfiles)
}

func TestProfileOverride_Verification(t *testing.T) {
	// Detailed verification with services
	viper.Reset()
	cmd := &cobra.Command{}
	BindFlags(cmd)
	os.Setenv("MCPANY_PROFILES", "env-profile")
	defer os.Unsetenv("MCPANY_PROFILES")

	fs := afero.NewMemMapFs()
	// Config defining profiles and services.
    // By default services are enabled unless disabled?
    // Manager logs "Checking service disabled status ... disabled_field false".
    // If not in profile override, it stays enabled.
    // So we need to DISABLE services by default in one profile and ENABLE in another?
    // Or assume default is enabled, so "file-profile" might disable "service-env"?

    // Let's use profile_definitions to control enablement.
	configContent := `
global_settings:
  profiles: ["file-profile"]
  profile_definitions:
    - name: "file-profile"
      service_config:
        service-file:
          enabled: true
        service-env:
          enabled: false
    - name: "env-profile"
      service_config:
        service-file:
          enabled: false
        service-env:
          enabled: true

upstream_services:
  - name: "service-env"
    http_service:
      address: "http://env"

  - name: "service-file"
    http_service:
      address: "http://file"
`
	_ = afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)

	gs := GlobalSettings()
	_ = gs.Load(cmd, fs)

	store := NewFileStore(fs, []string{"config.yaml"})
	cfg, err := LoadResolvedConfig(context.Background(), store)
	assert.NoError(t, err)

	// If env-profile wins, "service-env" should be enabled (present), "service-file" disabled (absent/skipped).

	services := cfg.GetUpstreamServices()

	foundEnv := false
	foundFile := false
	for _, s := range services {
		if s.GetName() == "service-env" { foundEnv = true }
		if s.GetName() == "service-file" { foundFile = true }
	}

	// Assert the CORRECT behavior: env wins
    // So service-env is found, service-file is NOT found.
	assert.False(t, foundFile, "Expected file-profile service NOT to be loaded (file-profile disabled it)")
	assert.True(t, foundEnv, "Expected env-profile service to be loaded (env-profile enabled it)")
}

func TestProfileDefaultFallback(t *testing.T) {
	viper.Reset()
	cmd := &cobra.Command{}
	BindFlags(cmd)
	// No env var set

	fs := afero.NewMemMapFs()
	// No profiles in config
	configContent := `
upstream_services: []
`
	_ = afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)

	store := NewFileStore(fs, []string{"config.yaml"})
	cfg, err := LoadResolvedConfig(context.Background(), store)
	assert.NoError(t, err)

	// Expect default
	assert.Equal(t, []string{"default"}, cfg.GetGlobalSettings().GetProfiles())
}
