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
				require.Len(t, cfg.GetUpstreamServices(), 1)
				assert.Equal(t, "service1", cfg.GetUpstreamServices()[0].GetName())
			},
		},
		{
			name:     "JSON Object (Single)",
			envValue: `{"name": "service2", "http_service": {"address": "http://example.org"}}`,
			verify: func(t *testing.T, cfg *configv1.McpAnyServerConfig) {
				require.Len(t, cfg.GetUpstreamServices(), 1)
				assert.Equal(t, "service2", cfg.GetUpstreamServices()[0].GetName())
			},
		},
		{
			name:     "Malformed JSON Array (Fallback to CSV)",
			envValue: `[{"name": "service3"`, // Invalid JSON
			verify: func(t *testing.T, cfg *configv1.McpAnyServerConfig) {
                // Should fall back to CSV processing.
                // The value `[{"name": "service3"` will be treated as a string and appended.
                // While this results in an invalid config structure (string instead of message),
                // we mainly want to ensure it doesn't crash the parser.
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

    require.Len(t, cfg.GetUpstreamServices(), 2)
    assert.Equal(t, "s1", cfg.GetUpstreamServices()[0].GetName())
    assert.Equal(t, "s2", cfg.GetUpstreamServices()[1].GetName())
}
