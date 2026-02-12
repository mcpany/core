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
	t.Run("Hardcoded Templates", func(t *testing.T) {
		store := memory.NewStore()
		tempDir, err := os.MkdirTemp("", "seed_test_empty")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		seeder := &Seeder{
			Store:       store,
			ExamplesDir: tempDir,
		}

		err = seeder.Seed(context.Background())
		require.NoError(t, err)

		// Verify templates are saved
		templates, err := store.ListServiceTemplates(context.Background())
		require.NoError(t, err)
		assert.NotEmpty(t, templates)

		// Check for some expected templates
		foundGithub := false
		foundGoogleCalendar := false

		for _, tmpl := range templates {
			if tmpl.GetId() == "github" {
				foundGithub = true
				assert.Equal(t, "GitHub", tmpl.GetName())
				assert.Contains(t, tmpl.GetTags(), "development")
			}
			if tmpl.GetId() == "google-calendar" {
				foundGoogleCalendar = true
				assert.Equal(t, "Google Calendar", tmpl.GetName())
			}
		}

		assert.True(t, foundGithub, "GitHub template not found")
		assert.True(t, foundGoogleCalendar, "Google Calendar template not found")
	})

	t.Run("File Scanning - Happy Path", func(t *testing.T) {
		store := memory.NewStore()
		tempDir, err := os.MkdirTemp("", "seed_test_happy")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create a mock service directory
		serviceDir := filepath.Join(tempDir, "mock-service")
		err = os.Mkdir(serviceDir, 0755)
		require.NoError(t, err)

		// Create a valid config.yaml
		configContent := `
upstream_services:
  - name: mock-service
    mcp_service:
      stdio_connection:
        command: "mock-cmd"
`
		err = os.WriteFile(filepath.Join(serviceDir, "config.yaml"), []byte(configContent), 0644)
		require.NoError(t, err)

		seeder := &Seeder{
			Store:       store,
			ExamplesDir: tempDir,
		}

		// It should run without error
		err = seeder.Seed(context.Background())
		require.NoError(t, err)

		// Note: Currently the file-based seeding logic doesn't save anything to the store,
		// so we just verify that it runs without crashing.
		// If implementation changes to actually save, we would assert existence here.
	})

	t.Run("File Scanning - Error Handling", func(t *testing.T) {
		store := memory.NewStore()
		tempDir, err := os.MkdirTemp("", "seed_test_error")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create a mock service directory with invalid YAML
		serviceDir := filepath.Join(tempDir, "bad-service")
		err = os.Mkdir(serviceDir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(serviceDir, "config.yaml"), []byte("invalid: yaml: : content"), 0644)
		require.NoError(t, err)

		seeder := &Seeder{
			Store:       store,
			ExamplesDir: tempDir,
		}

		// It should run without returning an error (it logs errors internally)
		err = seeder.Seed(context.Background())
		require.NoError(t, err)
	})
}
