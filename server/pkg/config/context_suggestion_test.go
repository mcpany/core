// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextAwareSuggestion_UpstreamService(t *testing.T) {
	yamlContent := `
version: v1
upstream_services:
  - name: "my-service"
    type: "http"
    http_config:
      address: "http://example.com"
`
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "config.yaml", []byte(yamlContent), 0644)

	store := NewFileStore(fs, []string{"config.yaml"})
	_, err := store.Load(context.Background())

	require.Error(t, err)
    // We expect it NOT to suggest tls_config if we fix it.
    // But currently it DOES suggest tls_config (as seen in my manual run).
    // So this test asserts that we DO NOT want tls_config.

    // Also, we want to see if we can get it to suggest http_service.
    // But given the distance, maybe not.
    // Let's at least ensure we don't get the confusing suggestion.

	assert.NotContains(t, err.Error(), "Did you mean \"tls_config\"?", "Should not suggest tls_config as it is not valid in this context")

    // Check for clean error message
    assert.Contains(t, err.Error(), "line 6: unknown field \"http_config\"")
    assert.NotContains(t, err.Error(), "proto: (line")
}

func TestContextAwareSuggestion_HttpService(t *testing.T) {
	yamlContent := `
version: v1
upstream_services:
  - name: "my-service"
    type: "http"
    http_service:
      addres: "http://example.com"
`
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "config.yaml", []byte(yamlContent), 0644)

	store := NewFileStore(fs, []string{"config.yaml"})
	_, err := store.Load(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Did you mean \"address\"?")
    assert.Contains(t, err.Error(), "line 7: unknown field \"addres\"")
}
