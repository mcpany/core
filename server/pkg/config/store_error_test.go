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

func TestYamlEngine_Unmarshal_LineNumber(t *testing.T) {
	yamlContent := `global_settings:
  mcp_listen_address: "0.0.0.0:50050"
  log_level: "INFO"
upstream_services:
  - name: "wttr.in"
    upstream_auth_typo: # Typo here (Line 6)
      api_key:
        param_name: "User-Agent"
        value:
          plain_text: "curl/8.5.0"
    http_service:
      address: "https://wttr.in"
`
	// Create a temp file
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "config.yaml", []byte(yamlContent), 0644)
	require.NoError(t, err)

	store := NewFileStore(fs, []string{"config.yaml"})
	_, err = store.Load(context.Background())

	require.Error(t, err)
	// Expect "line 6" in the error message
	t.Logf("Error message: %v", err)

    assert.Contains(t, err.Error(), "line 6")
    assert.Contains(t, err.Error(), "upstream_auth_typo")
}
