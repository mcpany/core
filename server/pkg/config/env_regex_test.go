/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestEnvVarRegexValidation(t *testing.T) {
	// Create a mock filesystem
	fs := afero.NewMemMapFs()
	// Use 'address' field which is a string and valid
	configContent := `
upstream_services:
  - name: "my-service"
    http_service:
      address: "${MCP_TEST_ADDR:regex=^https?://.*$}"
`
	afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)

	// Case 1: Valid Key
	os.Setenv("MCP_TEST_ADDR", "http://localhost:8080")
	t.Cleanup(func() { os.Unsetenv("MCP_TEST_ADDR") })

	store := NewFileStore(fs, []string{"config.yaml"})
	_, err := store.Load(context.Background())
	assert.NoError(t, err)

	// Case 2: Invalid Key
	os.Setenv("MCP_TEST_ADDR", "ftp://localhost:8080")
	_, err = store.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match required pattern")

	// Case 3: Missing Key (treated as missing because no default)
	os.Unsetenv("MCP_TEST_ADDR")
	_, err = store.Load(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable MCP_TEST_ADDR is missing")
}
