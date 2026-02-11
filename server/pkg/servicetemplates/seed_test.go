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
	// Create a temporary directory for examples
	tempDir, err := os.MkdirTemp("", "examples_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a mock store
	store := memory.NewStore()

	// Initialize Seeder
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Scenario 1: Hardcoded Templates (Success)
	// We expect the built-in templates to be seeded even if the directory is empty.
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify that templates were saved
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, templates, "Expected templates to be seeded")

	// Check specific templates (e.g., GitHub, Google Calendar)
	foundGithub := false
	foundCalendar := false
	for _, tmpl := range templates {
		if tmpl.GetId() == "github" {
			foundGithub = true
			assert.Equal(t, "GitHub", tmpl.GetName())
		}
		if tmpl.GetId() == "google-calendar" {
			foundCalendar = true
			assert.Equal(t, "Google Calendar", tmpl.GetName())
		}
	}
	assert.True(t, foundGithub, "Expected GitHub template")
	assert.True(t, foundCalendar, "Expected Google Calendar template")

	// Scenario 2: File Scanning (Happy Path - but currently no-op)
	// Create a valid example directory and config.yaml
	exampleDir := filepath.Join(tempDir, "my-service")
	require.NoError(t, os.Mkdir(exampleDir, 0755))

	configFile := filepath.Join(exampleDir, "config.yaml")
	configContent := `
upstream_services:
  - name: my-service
    upstream_auth:
      bearer_token:
        token:
          plain_text: "secret"
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	// Re-run Seed
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify that it didn't crash.
	// Note: Currently Seed() doesn't actually save file-based templates, so we can't assert that "my-service" exists in store.
	// But we can assert that existing templates are still there (idempotency/append).
	templates, err = store.ListServiceTemplates(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, templates)

	// Scenario 3: File Scanning (Invalid YAML)
	// Create another directory with invalid YAML
	invalidDir := filepath.Join(tempDir, "invalid-service")
	require.NoError(t, os.Mkdir(invalidDir, 0755))

	invalidConfigFile := filepath.Join(invalidDir, "config.yaml")
	require.NoError(t, os.WriteFile(invalidConfigFile, []byte("invalid: yaml: content: ["), 0644))

	// Re-run Seed
	// It should NOT return error, but log it (which we can't easily capture without redirecting stdout/log, but we verify it doesn't panic/error).
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify existing templates are still intact
	templates, err = store.ListServiceTemplates(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, templates)
}
