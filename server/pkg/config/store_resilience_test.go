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

func TestYamlEngine_ResilientPartialLoading(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Config with one valid service and one invalid service
	require.NoError(t, afero.WriteFile(fs, "/config/partial.yaml", []byte(`
upstream_services:
  - name: "valid-service"
    http_service:
      address: "https://valid.com"
  - name: "invalid-service"
    http_service:
      adress: "https://invalid.com" # Typo: adress -> address
`), 0644))

	store := NewFileStore(fs, []string{"/config/partial.yaml"})
	cfg, err := store.Load(context.Background())

	// We expect NO error from Load(), because it should have recovered.
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// We expect 1 service to be loaded
	require.Len(t, cfg.UpstreamServices, 1)
	assert.Equal(t, "valid-service", cfg.UpstreamServices[0].GetName())
}

func TestYamlEngine_ResilientPartialLoading_AllFail(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Config with only invalid services
	require.NoError(t, afero.WriteFile(fs, "/config/all_fail.yaml", []byte(`
upstream_services:
  - name: "invalid-service-1"
    http_service:
      adress: "https://example.com"
  - name: "invalid-service-2"
    http_service:
      adress: "https://example.org"
`), 0644))

	store := NewFileStore(fs, []string{"/config/all_fail.yaml"})
	_, err := store.Load(context.Background())

	// Should fail because no valid services were found
	assert.Error(t, err)
}

func TestYamlEngine_ResilientPartialLoading_GlobalError(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Config with global error (should fail completely)
	require.NoError(t, afero.WriteFile(fs, "/config/global_fail.yaml", []byte(`
global_settings:
  unknown_field: "bar"
upstream_services:
  - name: "valid-service"
    http_service:
      address: "https://valid.com"
`), 0644))

	store := NewFileStore(fs, []string{"/config/global_fail.yaml"})
	_, err := store.Load(context.Background())

	// Should fail because we can't salvage global settings
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown field")
}
