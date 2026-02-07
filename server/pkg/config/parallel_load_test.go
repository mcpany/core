// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_Load_Parallel(t *testing.T) {
	fs := afero.NewMemMapFs()

    // File 1: Sets API Key
	require.NoError(t, afero.WriteFile(fs, "/config/1.json", []byte(`{"global_settings": {"api_key": "key-1"}}`), 0644))
    // File 2: Overrides API Key (should win if processed in order)
	require.NoError(t, afero.WriteFile(fs, "/config/2.json", []byte(`{"global_settings": {"api_key": "key-2"}}`), 0644))
    // File 3: Adds a service
	require.NoError(t, afero.WriteFile(fs, "/config/3.yaml", []byte(`
upstream_services:
  - name: svc-3
    http_service:
      address: http://svc-3
`), 0644))

	store := NewFileStore(fs, []string{"/config"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

    // Check precedence: 2.json should override 1.json
	assert.Equal(t, "key-2", cfg.GetGlobalSettings().GetApiKey())
    // Check merging: svc-3 should exist
    require.Len(t, cfg.GetUpstreamServices(), 1)
    assert.Equal(t, "svc-3", cfg.GetUpstreamServices()[0].GetName())
}

func TestFileStore_Load_Parallel_Error(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/config/1.json", []byte(`{}`), 0644))
	require.NoError(t, afero.WriteFile(fs, "/config/bad.json", []byte(`{invalid json`), 0644))

	store := NewFileStore(fs, []string{"/config"})
	_, err := store.Load(context.Background())
	assert.Error(t, err)
    // Should fail on bad.json
}

func TestFileStore_Load_Parallel_SkipError(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/config/1.json", []byte(`{"global_settings": {"api_key": "key-1"}}`), 0644))
	require.NoError(t, afero.WriteFile(fs, "/config/bad.json", []byte(`{invalid json`), 0644))

	store := NewFileStoreWithSkipErrors(fs, []string{"/config"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

    // Should still have 1.json applied
    assert.Equal(t, "key-1", cfg.GetGlobalSettings().GetApiKey())
}
