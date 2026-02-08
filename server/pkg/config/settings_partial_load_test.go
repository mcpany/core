package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSettings_Load_PartialConfigFailure validates that global settings (like mcp-listen-address)
// are loaded from the config file even if there are validation errors in the services.
// This ensures that the server attempts to bind to the correct address even if services are misconfigured.
func TestSettings_Load_PartialConfigFailure(t *testing.T) {
	viper.Reset()
	fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}

	viper.Set("config-path", []string{"/config.yaml"})

	// Create a config file with:
	// 1. Custom mcp_listen_address (50055)
	// 2. A VALID service (to show it's parsed)
	// 3. An INVALID service (missing address) which passes schema validation but fails logical validation
	configContent := `
global_settings:
  mcp_listen_address: "127.0.0.1:50055"

upstream_services:
  - id: valid-service
    http_service:
      address: "http://example.com"

  - id: invalid-service
    http_service:
      # Missing address, which causes Validate (manual) to fail
      tools: []
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}

	// We expect Load to NOT return an error here because Settings.Load
	// swallows the error when peeking for listen address.
	// But we want to verify that it DID pick up the listen address.
	err = settings.Load(cmd, fs)
	require.NoError(t, err)

	// verification: The listen address should be the one from the config file (50055),
	// not the default (empty or 50050).
	assert.Equal(t, "127.0.0.1:50055", settings.MCPListenAddress())
}
