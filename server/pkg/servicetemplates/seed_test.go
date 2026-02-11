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
	// Create a temporary directory for templates
	tempDir, err := os.MkdirTemp("", "templates")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create subdirectories and files
	validDir := filepath.Join(tempDir, "valid")
	require.NoError(t, os.Mkdir(validDir, 0755))
	validConfig := `
upstream_services:
  - name: valid-service
    type: object
`
	require.NoError(t, os.WriteFile(filepath.Join(validDir, "config.yaml"), []byte(validConfig), 0644))

	invalidDir := filepath.Join(tempDir, "invalid")
	require.NoError(t, os.Mkdir(invalidDir, 0755))
	invalidConfig := `
upstream_services:
  - name: invalid-service
    : invalid syntax
`
	require.NoError(t, os.WriteFile(filepath.Join(invalidDir, "config.yaml"), []byte(invalidConfig), 0644))

	missingConfigDir := filepath.Join(tempDir, "missing")
	require.NoError(t, os.Mkdir(missingConfigDir, 0755))

	// Initialize the store
	store := memory.NewStore()

	// Initialize the seeder
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Run Seed
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify that the store contains the built-in templates
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)

	// We expect at least the built-in templates (7 at the time of writing)
	assert.GreaterOrEqual(t, len(templates), 7)

	// Verify specific templates are present
	expectedIDs := []string{
		"google-calendar",
		"github",
		"gitlab",
		"slack",
		"notion",
		"linear",
		"jira",
	}

	foundIDs := make(map[string]bool)
	for _, t := range templates {
		foundIDs[t.GetId()] = true
	}

	for _, id := range expectedIDs {
		assert.True(t, foundIDs[id], "Expected template %s to be present", id)
	}

	// Verify that the file-based scanning didn't crash
	// (Note: Currently the file scanning logic doesn't save anything, so we can't assert on "valid-service" being present)
	// But passing `require.NoError(t, err)` confirms it handled errors gracefully.
}
