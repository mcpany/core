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

func TestFileStore_Load_SetsReadOnly(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/config/service.yaml", []byte(`
upstream_services:
  - name: test-service
    http_service:
      address: http://example.com
`), 0644))

	store := NewFileStore(fs, []string{"/config/service.yaml"})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)

	require.Len(t, cfg.GetUpstreamServices(), 1)
	svc := cfg.GetUpstreamServices()[0]
	assert.Equal(t, "test-service", svc.GetName())
	assert.True(t, svc.GetReadOnly(), "Service loaded from file should be ReadOnly")
}
