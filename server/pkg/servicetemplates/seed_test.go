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

func TestSeeder_Seed_BuiltIn(t *testing.T) {
	// Setup
	store := memory.NewStore()
	seeder := &servicetemplates.Seeder{
		Store:       store,
		ExamplesDir: t.TempDir(), // Empty directory
	}

	// Execution
	err := seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verification
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)

	// Check that we have at least the built-in templates
	// Currently there are 7 built-in templates: google-calendar, github, gitlab, slack, notion, linear, jira
	assert.GreaterOrEqual(t, len(templates), 7)

	expectedIDs := []string{"google-calendar", "github", "gitlab", "slack", "notion", "linear", "jira"}
	foundIDs := make(map[string]bool)
	for _, t := range templates {
		foundIDs[t.GetId()] = true
	}

	for _, id := range expectedIDs {
		assert.True(t, foundIDs[id], "Template %s should be present", id)
	}
}

func TestSeeder_Seed_FileScan_Valid(t *testing.T) {
	// Setup
	store := memory.NewStore()
	tempDir := t.TempDir()

	// Create a valid example structure
	// server/examples/my-service/config.yaml
	serviceDir := filepath.Join(tempDir, "my-service")
	err := os.Mkdir(serviceDir, 0755)
	require.NoError(t, err)

	configContent := `
upstream_services:
- name: my-service
  mcp_service:
    stdio_connection:
      command: "echo"
      args: ["hello"]
`
	err = os.WriteFile(filepath.Join(serviceDir, "config.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	seeder := &servicetemplates.Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Execution
	// Currently the file scanning logic does not save anything to the store,
	// but it should run without error.
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verification
	// Still check built-ins are there
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(templates), 7)
}

func TestSeeder_Seed_FileScan_InvalidYAML(t *testing.T) {
	// Setup
	store := memory.NewStore()
	tempDir := t.TempDir()

	// Create an invalid example structure
	serviceDir := filepath.Join(tempDir, "broken-service")
	err := os.Mkdir(serviceDir, 0755)
	require.NoError(t, err)

	configContent := `
upstream_services:
- name: broken-service
  mcp_service:
    stdio_connection:
      command: "echo"
      args: ["hello"  <-- syntax error
`
	err = os.WriteFile(filepath.Join(serviceDir, "config.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	seeder := &servicetemplates.Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Execution
	// The seeder logs errors but continues (does not return error).
	err = seeder.Seed(context.Background())
	require.NoError(t, err)
}

func TestSeeder_Seed_FileScan_MissingConfig(t *testing.T) {
	// Setup
	store := memory.NewStore()
	tempDir := t.TempDir()

	// Create a directory without config.yaml
	serviceDir := filepath.Join(tempDir, "empty-service")
	err := os.Mkdir(serviceDir, 0755)
	require.NoError(t, err)

	seeder := &servicetemplates.Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Execution
	err = seeder.Seed(context.Background())
	require.NoError(t, err)
}
