// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestMissingEnvVar(t *testing.T) {
	// Create a mock filesystem
	fs := afero.NewMemMapFs()
	configContent := `
upstream_services:
  - name: "my-service"
    http_service:
      address: "http://localhost:${MCP_TEST_MISSING_VAR}"
`
	afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)

	// Ensure variable is not set
	val, present := os.LookupEnv("MCP_TEST_MISSING_VAR")
	if present {
		t.Cleanup(func() { os.Setenv("MCP_TEST_MISSING_VAR", val) })
	} else {
		t.Cleanup(func() { os.Unsetenv("MCP_TEST_MISSING_VAR") })
	}
	os.Unsetenv("MCP_TEST_MISSING_VAR")

	store := NewFileStore(fs, []string{"config.yaml"})
	_, err := store.Load(context.Background())

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Line 5: variable ${MCP_TEST_MISSING_VAR} is missing")
		assert.Contains(t, err.Error(), "Fix: Set these environment variables")
	}
}
