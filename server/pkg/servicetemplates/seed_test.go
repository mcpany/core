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
	tempDir := t.TempDir()

	// Create a valid config.yaml
	validDir := filepath.Join(tempDir, "valid-service")
	require.NoError(t, os.Mkdir(validDir, 0755))
	validConfig := `
upstream_services:
- name: valid-service
  upstream_auth:
    bearer_token:
      token:
        plain_text: "${TOKEN}"
`
	require.NoError(t, os.WriteFile(filepath.Join(validDir, "config.yaml"), []byte(validConfig), 0644))

	// Create an invalid config.yaml
	invalidDir := filepath.Join(tempDir, "invalid-service")
	require.NoError(t, os.Mkdir(invalidDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(invalidDir, "config.yaml"), []byte("invalid: yaml: :"), 0644))

	// Create a directory without config.yaml
	emptyDir := filepath.Join(tempDir, "empty-dir")
	require.NoError(t, os.Mkdir(emptyDir, 0755))

	store := memory.NewStore()
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	err := seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify hardcoded templates are present
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)

	// Check for a few expected templates
	expectedTemplates := map[string]bool{
		"github":          false,
		"google-calendar": false,
		"slack":           false,
		"jira":            false,
		"gitlab":          false,
	}

	for _, tmpl := range templates {
		if _, expected := expectedTemplates[tmpl.GetId()]; expected {
			expectedTemplates[tmpl.GetId()] = true
		}
	}

	for id, found := range expectedTemplates {
		assert.True(t, found, "Template %s should be present", id)
	}

	// Verify that the seed logic didn't crash on invalid YAML or missing files.
	// Note: Currently the file scanning logic doesn't save anything, so we can't assert on "valid-service" being present.
	// But ensuring it ran without error is valuable.
}

func TestSeeder_Seed_StoreError(t *testing.T) {
	// Since we are using memory store which doesn't really error on Save,
	// checking store errors is hard without mocking the store interface.
	// However, we can check basic initialization.
	store := memory.NewStore()
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: "non-existent-dir",
	}

	// Should fail because directory does not exist
	err := seeder.Seed(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read examples dir")
}
