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
	tempDir, err := os.MkdirTemp("", "servicetemplates_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a valid example directory and config.yaml
	validDir := filepath.Join(tempDir, "valid-service")
	require.NoError(t, os.Mkdir(validDir, 0755))
	validConfig := `
upstream_services:
- name: valid-service
  upstream_auth:
    bearer_token:
      token:
        plain_text: "${VALID_TOKEN}"
`
	require.NoError(t, os.WriteFile(filepath.Join(validDir, "config.yaml"), []byte(validConfig), 0644))

	// Create an invalid example directory (bad yaml)
	invalidDir := filepath.Join(tempDir, "invalid-service")
	require.NoError(t, os.Mkdir(invalidDir, 0755))
	invalidConfig := `
upstream_services:
- name: invalid-service
  upstream_auth:
    bearer_token:
      token: [ broken yaml
`
	require.NoError(t, os.WriteFile(filepath.Join(invalidDir, "config.yaml"), []byte(invalidConfig), 0644))

	// Create a directory without config.yaml (should be ignored)
	emptyDir := filepath.Join(tempDir, "empty-service")
	require.NoError(t, os.Mkdir(emptyDir, 0755))

	// Initialize store and seeder
	store := memory.NewStore()
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Run Seed
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify hardcoded templates are present
	// We expect "google-calendar", "github", "gitlab", "slack", "notion", "linear", "jira"
	expectedIDs := []string{"google-calendar", "github", "gitlab", "slack", "notion", "linear", "jira"}

	for _, id := range expectedIDs {
		tmpl, err := store.GetServiceTemplate(context.Background(), id)
		require.NoError(t, err)
		assert.NotNil(t, tmpl, "Template %s should be present", id)
		assert.Equal(t, id, tmpl.GetId())
	}

	// Verify that the valid service from file was NOT added (as per current implementation)
	// If the implementation changes to actually add them, this test should be updated.
	// Currently, the implementation parses but doesn't store file-based templates.
	tmpl, err := store.GetServiceTemplate(context.Background(), "valid-service")
	require.NoError(t, err)
	assert.Nil(t, tmpl, "File-based template should not be present (yet)")

	// Verify no panic or error occurred despite invalid yaml file
}
