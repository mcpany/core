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
	tempDir, err := os.MkdirTemp("", "examples")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create valid example directory structure
	validServiceDir := filepath.Join(tempDir, "test-service")
	err = os.Mkdir(validServiceDir, 0755)
	require.NoError(t, err)

	validConfig := `
upstream_services:
  - name: test-service
    mcp_service:
      stdio_connection:
        command: "test"
`
	err = os.WriteFile(filepath.Join(validServiceDir, "config.yaml"), []byte(validConfig), 0644)
	require.NoError(t, err)

	// Create invalid YAML example
	invalidServiceDir := filepath.Join(tempDir, "invalid-service")
	err = os.Mkdir(invalidServiceDir, 0755)
	require.NoError(t, err)

	invalidConfig := `
upstream_services:
  - name: invalid-service
    INVALID_YAML
`
	err = os.WriteFile(filepath.Join(invalidServiceDir, "config.yaml"), []byte(invalidConfig), 0644)
	require.NoError(t, err)

	// Create empty example (no config)
	emptyServiceDir := filepath.Join(tempDir, "empty-service")
	err = os.Mkdir(emptyServiceDir, 0755)
	require.NoError(t, err)

	// Setup Seeder with Memory Store
	store := memory.NewStore()
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Execute Seed
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify Hardcoded Templates are present
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)

	// Check for a few expected hardcoded templates
	expectedTemplates := map[string]bool{
		"google-calendar": false,
		"github":          false,
		"gitlab":          false,
		"slack":           false,
		"notion":          false,
		"linear":          false,
		"jira":            false,
	}

	for _, tmpl := range templates {
		if _, ok := expectedTemplates[tmpl.GetId()]; ok {
			expectedTemplates[tmpl.GetId()] = true
		}
	}

	for id, found := range expectedTemplates {
		assert.True(t, found, "Expected template %s to be seeded", id)
	}

	// Verify that the function didn't crash on invalid input (implicit by reaching here)
	// Currently, the file-based seeding is a no-op, so we don't expect "test-service" to be in templates.
	// If/when the file-based logic is implemented, this assertion can be updated.
	// For now, we just ensure it didn't break execution.
}
