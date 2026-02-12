// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package servicetemplates_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/servicetemplates"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeed(t *testing.T) {
	// Create a temporary directory for examples
	tempDir, err := os.MkdirTemp("", "examples")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a mock store
	store := memory.NewStore()

	// Initialize Seeder
	seeder := &servicetemplates.Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	t.Run("Hardcoded Templates", func(t *testing.T) {
		// Run Seed
		err := seeder.Seed(context.Background())
		require.NoError(t, err)

		// Verify hardcoded templates are present
		// We expect: google-calendar, github, gitlab, slack, notion, linear, jira
		expectedIDs := []string{
			"google-calendar",
			"github",
			"gitlab",
			"slack",
			"notion",
			"linear",
			"jira",
		}

		for _, id := range expectedIDs {
			tmpl, err := store.GetServiceTemplate(context.Background(), id)
			require.NoError(t, err)
			require.NotNil(t, tmpl, "Template %s should be present", id)
			assert.Equal(t, id, tmpl.GetId())
		}
	})

	t.Run("File Scanning Happy Path", func(t *testing.T) {
		// Create a valid service directory
		svcDir := filepath.Join(tempDir, "custom-service")
		err := os.Mkdir(svcDir, 0755)
		require.NoError(t, err)

		// Create a valid config.yaml
		configPath := filepath.Join(svcDir, "config.yaml")
		configContent := `
upstream_services:
  - name: custom-service
    upstream_auth:
      bearer_token:
        token: "secret"
`
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Run Seed again (it should re-seed hardcoded ones and try to parse the file)
		// Since the file parsing logic is currently a no-op for saving, we just verify it doesn't crash.
		err = seeder.Seed(context.Background())
		require.NoError(t, err)
	})

	t.Run("File Scanning Invalid YAML", func(t *testing.T) {
		// Create a directory with invalid YAML
		svcDir := filepath.Join(tempDir, "invalid-yaml")
		err := os.Mkdir(svcDir, 0755)
		require.NoError(t, err)

		configPath := filepath.Join(svcDir, "config.yaml")
		// Invalid YAML: missing value
		configContent := `
upstream_services:
  - name: invalid
    upstream_auth:
`
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Run Seed
		// It should print an error (captured if we redirected stdout, but here we just check it doesn't return error)
		err = seeder.Seed(context.Background())
		require.NoError(t, err)
	})

	t.Run("File Scanning No Config", func(t *testing.T) {
		// Create a directory without config.yaml
		svcDir := filepath.Join(tempDir, "no-config")
		err := os.Mkdir(svcDir, 0755)
		require.NoError(t, err)

		// Run Seed
		err = seeder.Seed(context.Background())
		require.NoError(t, err)
	})
}
