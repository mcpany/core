// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadServices_Coverage(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/config.yaml", []byte("global_settings:\n  log_level: INFO"), 0644))
	store := NewFileStore(fs, []string{"/config.yaml"})
	ctx := context.Background()

	// Invalid binary type
	_, err := LoadServices(ctx, store, "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown binary type")

	// Worker binary type
	cfg, err := LoadServices(ctx, store, "worker")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestLoadServices_ValidationErrors(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Config with invalid service (empty name)
	require.NoError(t, afero.WriteFile(fs, "/config.yaml", []byte(`
upstream_services:
  - http_service:
      address: http://example.com
`), 0644))
	store := NewFileStore(fs, []string{"/config.yaml"})
	ctx := context.Background()

	_, err := LoadServices(ctx, store, "server")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Configuration Validation Failed")
}

func TestFileStore_Load_Coverage(t *testing.T) {
	// Skip Errors
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/bad.yaml", []byte(":"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/good.yaml", []byte("global_settings:\n  log_level: INFO"), 0644))

	store := NewFileStoreWithSkipErrors(fs, []string{"/bad.yaml", "/good.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_INFO, cfg.GetGlobalSettings().GetLogLevel())

	// Expand Error
	require.NoError(t, afero.WriteFile(fs, "/expand.yaml", []byte("key: ${MISSING}"), 0644))
	store = NewFileStore(fs, []string{"/expand.yaml"})
	_, err = store.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing environment variables")
}
