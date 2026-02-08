package config

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_MergeStrategy(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Base config with one service and one profile
	require.NoError(t, afero.WriteFile(fs, "/config/01-base.yaml", []byte(`
global_settings:
  profiles: ["default"]
upstream_services:
  - name: "base-service"
    http_service: { address: "http://base" }
`), 0644))

	// Overlay config with replace strategy
	require.NoError(t, afero.WriteFile(fs, "/config/02-replace.yaml", []byte(`
merge_strategy:
  upstream_service_list: "replace"
  profile_list: "replace"

global_settings:
  profiles: ["prod"]
upstream_services:
  - name: "overlay-service"
    http_service: { address: "http://overlay" }
`), 0644))

	store := NewFileStore(fs, []string{"/config/01-base.yaml", "/config/02-replace.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	// Verify upstream services list was replaced
	require.Len(t, cfg.GetUpstreamServices(), 1)
	assert.Equal(t, "overlay-service", cfg.GetUpstreamServices()[0].GetName())

	// Verify profiles list was replaced
	require.Len(t, cfg.GetGlobalSettings().GetProfiles(), 1)
	assert.Equal(t, "prod", cfg.GetGlobalSettings().GetProfiles()[0])
}

func TestFileStore_MergeStrategy_DefaultExtend(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Base config
	require.NoError(t, afero.WriteFile(fs, "/config/01-base.yaml", []byte(`
upstream_services:
  - name: "base-service"
    http_service: { address: "http://base" }
`), 0644))

	// Overlay config (default extend)
	require.NoError(t, afero.WriteFile(fs, "/config/02-extend.yaml", []byte(`
upstream_services:
  - name: "overlay-service"
    http_service: { address: "http://overlay" }
`), 0644))

	store := NewFileStore(fs, []string{"/config/01-base.yaml", "/config/02-extend.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	// Verify upstream services list was extended
	require.Len(t, cfg.GetUpstreamServices(), 2)
	names := []string{cfg.GetUpstreamServices()[0].GetName(), cfg.GetUpstreamServices()[1].GetName()}
	assert.Contains(t, names, "base-service")
	assert.Contains(t, names, "overlay-service")
}
