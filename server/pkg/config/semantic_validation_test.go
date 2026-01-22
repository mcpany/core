package config

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_Load_SemanticValidation_LineNumber(t *testing.T) {
	// A config that is valid YAML but invalid per "validateUpstreamService"
	// because it lacks a service type (http_service, etc.)
	yamlContent := `global_settings:
  mcp_listen_address: "0.0.0.0:50050"
upstream_services:
  - name: "valid-service"
    http_service:
      address: "http://localhost"

  - name: "invalid-service"
    # Missing service config (http_service, etc.)
`
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "config.yaml", []byte(yamlContent), 0644)
	require.NoError(t, err)

	store := NewFileStore(fs, []string{"config.yaml"})
	_, err = store.Load(context.Background())

	require.Error(t, err)
	t.Logf("Error message: %v", err)

	// This assertion verifies the fix
	assert.Contains(t, err.Error(), "line 8")
	assert.Contains(t, err.Error(), "invalid-service")
	assert.Contains(t, err.Error(), "service type not specified")
}
