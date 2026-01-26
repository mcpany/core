// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvVar_RepeatedMessage_WithComma verifies that we can set repeated message fields
// via environment variables even if the JSON objects contain commas.
// This was a bug where the CSV parser would split the JSON object at the internal comma.
func TestEnvVar_RepeatedMessage_WithComma(t *testing.T) {
	// Setup env var with repeated message containing comma
	// MCPANY__UPSTREAM_SERVICES maps to upstream_services
	os.Setenv("MCPANY__UPSTREAM_SERVICES", `{"name": "service,1", "http_service": {"address": "http://localhost:8080"}}, {"name": "service2", "http_service": {"address": "http://localhost:8081"}}`)
	defer os.Unsetenv("MCPANY__UPSTREAM_SERVICES")

	fs := afero.NewMemMapFs()
	// Non-empty config file to trigger loading
	afero.WriteFile(fs, "config.yaml", []byte("global_settings: {}"), 0644)

	store := NewFileStore(fs, []string{"config.yaml"})
	cfg, err := store.Load(context.Background())

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 2, len(cfg.GetUpstreamServices()))
	if len(cfg.GetUpstreamServices()) >= 1 {
		assert.Equal(t, "service,1", cfg.GetUpstreamServices()[0].GetName())
	}
}
