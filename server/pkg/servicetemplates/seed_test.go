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
	// Setup temporary directory for examples
	tempDir, err := os.MkdirTemp("", "servicetemplates_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a dummy config.yaml in a subdirectory
	exampleServiceDir := filepath.Join(tempDir, "example-service")
	err = os.Mkdir(exampleServiceDir, 0755)
	require.NoError(t, err)

	configContent := `
upstream_services:
- name: example-service
  upstream_auth:
    bearer_token:
      token:
        plain_text: "token"
`
	err = os.WriteFile(filepath.Join(exampleServiceDir, "config.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	// Create a store
	store := memory.NewStore()

	// Initialize Seeder
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Run Seed
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify Hardcoded Templates
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)

	// We expect the hardcoded templates to be present.
	// These IDs should match the ones defined in seed.go.
	expectedIDs := []string{"google-calendar", "github", "gitlab", "slack", "notion", "linear", "jira"}

	foundIDs := make(map[string]bool)
	for _, tmpl := range templates {
		foundIDs[tmpl.GetId()] = true
	}

	for _, id := range expectedIDs {
		assert.True(t, foundIDs[id], "Template %s should be seeded", id)
	}

	// Verify File Scanning Logic (Currently it doesn't save anything, but we ensure it runs without error)
	// If the logic changes to actually save the file-based templates, this test should be updated.
	// For now, we just assert that NO error occurred during Seed, which implies the file scanning didn't crash.
}

func TestSeeder_Seed_InvalidYAML(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "servicetemplates_test_invalid")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a subdirectory with invalid YAML
	invalidServiceDir := filepath.Join(tempDir, "invalid-service")
	err = os.Mkdir(invalidServiceDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(invalidServiceDir, "config.yaml"), []byte("invalid: yaml: :"), 0644)
	require.NoError(t, err)

	store := memory.NewStore()
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Seed should not return error, it just logs and continues
	err = seeder.Seed(context.Background())
	assert.NoError(t, err)

	// Verify that hardcoded templates are still seeded
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, templates)
}

func TestSeeder_Seed_NoExamplesDir(t *testing.T) {
	store := memory.NewStore()
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: "/path/that/does/not/exist",
	}

	// Seed should return error because it can't read the directory
	err := seeder.Seed(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read examples dir")
}
