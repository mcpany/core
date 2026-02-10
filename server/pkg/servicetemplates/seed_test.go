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

func TestSeeder_Seed_Hardcoded(t *testing.T) {
	// Create a mock store
	store := memory.NewStore()

	// Create a temporary directory for examples (empty for now)
	tempDir, err := os.MkdirTemp("", "examples_empty")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Seed the database
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify hardcoded templates are present
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, templates)

	// Check for a specific template (e.g., GitHub)
	foundGithub := false
	for _, tmpl := range templates {
		if tmpl.GetId() == "github" {
			foundGithub = true
			assert.Equal(t, "GitHub", tmpl.GetName())
			assert.Contains(t, tmpl.GetTags(), "development")
			break
		}
	}
	assert.True(t, foundGithub, "GitHub template should be seeded")
}

func TestSeeder_Seed_FileScanning_Valid(t *testing.T) {
	// Create a mock store
	store := memory.NewStore()

	// Create a temporary directory structure mimicking server/examples
	tempDir, err := os.MkdirTemp("", "examples_valid")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a valid service directory
	serviceDir := filepath.Join(tempDir, "test-service")
	err = os.Mkdir(serviceDir, 0755)
	require.NoError(t, err)

	// Create a valid config.yaml
	configContent := `
upstream_services:
- name: test-service
  mcp_service:
    http_connection:
      http_address: "http://localhost:8080"
`
	err = os.WriteFile(filepath.Join(serviceDir, "config.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Run Seed (should not crash)
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify hardcoded templates are still present
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, templates)

	// Note: The current implementation of Seed ignores the file content (does not persist it).
	// So we don't expect "test-service" to be in the store yet.
	// This test ensures that the scanning logic runs without error.
}

func TestSeeder_Seed_FileScanning_InvalidYAML(t *testing.T) {
	// Create a mock store
	store := memory.NewStore()

	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "examples_invalid")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a service directory
	serviceDir := filepath.Join(tempDir, "bad-service")
	err = os.Mkdir(serviceDir, 0755)
	require.NoError(t, err)

	// Create an invalid config.yaml
	err = os.WriteFile(filepath.Join(serviceDir, "config.yaml"), []byte("invalid: yaml: ["), 0644)
	require.NoError(t, err)

	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Run Seed (should log error but not fail)
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify store is still functional (hardcoded templates loaded)
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, templates)
}

func TestSeeder_Seed_ExamplesDir_NotDir(t *testing.T) {
	// Create a mock store
	store := memory.NewStore()

	// Use a file as ExamplesDir (should fail)
	tempFile := filepath.Join(os.TempDir(), "not_a_dir")
	f, err := os.Create(tempFile)
	require.NoError(t, err)
	f.Close()
	defer os.Remove(tempFile)

	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempFile,
	}

	// Run Seed (should fail because ReadDir fails)
	err = seeder.Seed(context.Background())
	assert.Error(t, err)
}

func TestSeeder_Seed_SubDir_NoConfig(t *testing.T) {
	// Create a mock store
	store := memory.NewStore()

	tempDir, err := os.MkdirTemp("", "examples_noconfig")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a directory without config.yaml
	err = os.Mkdir(filepath.Join(tempDir, "empty-dir"), 0755)
	require.NoError(t, err)

	// Create a file in root (should be ignored)
	err = os.WriteFile(filepath.Join(tempDir, "ignored.txt"), []byte(""), 0644)
	require.NoError(t, err)

	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	err = seeder.Seed(context.Background())
	require.NoError(t, err)
}
