// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package servicetemplates

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeeder_Seed(t *testing.T) {
	// Setup
	store := memory.NewStore()
	examplesDir := t.TempDir()
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: examplesDir,
	}

	// Create test data
	// 1. Valid config
	validDir := filepath.Join(examplesDir, "valid-service")
	require.NoError(t, os.Mkdir(validDir, 0755))
	validConfig := `
upstream_services:
  - name: valid-service
    mcp_service:
      http_connection:
        http_address: "http://localhost:8080"
`
	require.NoError(t, os.WriteFile(filepath.Join(validDir, "config.yaml"), []byte(validConfig), 0644))

	// 2. Invalid YAML
	invalidDir := filepath.Join(examplesDir, "invalid-service")
	require.NoError(t, os.Mkdir(invalidDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(invalidDir, "config.yaml"), []byte("invalid: [yaml"), 0644))

	// 3. No config
	emptyDir := filepath.Join(examplesDir, "empty-service")
	require.NoError(t, os.Mkdir(emptyDir, 0755))

	// Execute
	err := seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify Hardcoded Templates
	expectedTemplates := []string{
		"google-calendar",
		"github",
		"gitlab",
		"slack",
		"notion",
		"linear",
		"jira",
	}

	for _, id := range expectedTemplates {
		tmpl, err := store.GetServiceTemplate(context.Background(), id)
		require.NoError(t, err, "Failed to get template %s", id)
		assert.NotNil(t, tmpl, "Template %s should exist", id)
		assert.Equal(t, id, tmpl.GetId())
	}

	// Verify File Scanning behavior
	// Currently, the file scanning logic does NOT save templates to the store.
	// So we verify that it didn't crash and didn't save the valid service.
	// If the logic is updated later to actually save them, this test should be updated.
	tmpl, err := store.GetServiceTemplate(context.Background(), "valid-service")
	require.NoError(t, err) // Store returns nil, nil if not found
	assert.Nil(t, tmpl, "valid-service should NOT be saved (current behavior)")
}
