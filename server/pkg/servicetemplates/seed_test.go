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
		// Create mock store
		store := memory.NewStore()

		// Create temp dir for examples (empty for now)
		tempDir := t.TempDir()

		seeder := Seeder{
			Store:       store,
			ExamplesDir: tempDir,
		}

		err := seeder.Seed(context.Background())
		require.NoError(t, err)

		// Verify hardcoded templates are present
		// We expect "github", "google-calendar", etc.
		// Let's check "github" specifically as per plan
		tmpl, err := store.GetServiceTemplate(context.Background(), "github")
		require.NoError(t, err)
		require.NotNil(t, tmpl)
		assert.Equal(t, "github", tmpl.GetId())
		assert.Equal(t, "GitHub", tmpl.GetName())
	})

	t.Run("File Scanning - Happy Path", func(t *testing.T) {
		// Create mock store
		store := memory.NewStore()

		// Create temp dir
		tempDir := t.TempDir()

		// Create a mock service directory and config
		svcDir := filepath.Join(tempDir, "mock-service")
		err := os.Mkdir(svcDir, 0755)
		require.NoError(t, err)

		configContent := `
upstream_services:
  - name: mock-service
    mcp_service:
      stdio_connection:
        command: "echo"
        args: ["hello"]
`
		err = os.WriteFile(filepath.Join(svcDir, "config.yaml"), []byte(configContent), 0644)
		require.NoError(t, err)

		seeder := Seeder{
			Store:       store,
			ExamplesDir: tempDir,
		}

		// Run seed. Currently, the file scanning logic parses but does not save.
		// So we just verify it doesn't crash.
		err = seeder.Seed(context.Background())
		require.NoError(t, err)

		// Verify hardcoded templates are still there
		tmpl, err := store.GetServiceTemplate(context.Background(), "github")
		require.NoError(t, err)
		require.NotNil(t, tmpl)
	})

	t.Run("File Scanning - Invalid YAML", func(t *testing.T) {
		// Create mock store
		store := memory.NewStore()

		// Create temp dir
		tempDir := t.TempDir()

		// Create a mock service directory with invalid config
		svcDir := filepath.Join(tempDir, "bad-service")
		err := os.Mkdir(svcDir, 0755)
		require.NoError(t, err)

		// Invalid YAML
		err = os.WriteFile(filepath.Join(svcDir, "config.yaml"), []byte(": - invalid yaml"), 0644)
		require.NoError(t, err)

		seeder := Seeder{
			Store:       store,
			ExamplesDir: tempDir,
		}

		// Should not fail, just log error and continue
		err = seeder.Seed(context.Background())
		require.NoError(t, err)

		// Verify hardcoded templates are still there
		tmpl, err := store.GetServiceTemplate(context.Background(), "github")
		require.NoError(t, err)
		require.NotNil(t, tmpl)
	})

	t.Run("File Scanning - Invalid Directory", func(t *testing.T) {
		store := memory.NewStore()
		seeder := Seeder{
			Store:       store,
			ExamplesDir: "/path/to/non/existent/directory",
		}

		// Should fail because ReadDir fails
		err := seeder.Seed(context.Background())
		assert.Error(t, err)
	})
}
