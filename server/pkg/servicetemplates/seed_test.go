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
	// t.TempDir() automatically cleans up after the test
	tempDir := t.TempDir()

	// Create a dummy service directory and config
	dummyServiceDir := filepath.Join(tempDir, "dummy-service")
	err := os.Mkdir(dummyServiceDir, 0755)
	require.NoError(t, err)

	dummyConfig := `
upstream_services:
- name: dummy-service
  description: "A dummy service for testing"
`
	err = os.WriteFile(filepath.Join(dummyServiceDir, "config.yaml"), []byte(dummyConfig), 0644)
	require.NoError(t, err)

	// Create another dummy service with invalid YAML to test error handling
	invalidServiceDir := filepath.Join(tempDir, "invalid-service")
	err = os.Mkdir(invalidServiceDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(invalidServiceDir, "config.yaml"), []byte("invalid: yaml: ["), 0644)
	require.NoError(t, err)

	// specific valid service from hardcoded list
	// We don't need to create files for hardcoded ones as they are... hardcoded.

	// Initialize Store
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

	// Check that we have at least the hardcoded ones
	// The current implementation has 7 hardcoded templates:
	// google-calendar, github, gitlab, slack, notion, linear, jira
	expectedIDs := []string{
		"google-calendar",
		"github",
		"gitlab",
		"slack",
		"notion",
		"linear",
		"jira",
	}

	foundIDs := make(map[string]bool)
	for _, t := range templates {
		foundIDs[t.GetId()] = true
	}

	for _, id := range expectedIDs {
		assert.True(t, foundIDs[id], "Expected template %s not found", id)
	}

	// Verify that file-based templates are NOT added (as per current implementation behavior)
	// The dummy-service should NOT be in the templates list because the code just parses and ignores it.
	// This assertion confirms the current behavior (and coverage of that code path).
	assert.False(t, foundIDs["dummy-service"], "Dummy service should not be added (yet)")
}

func TestSeeder_Seed_InvalidDir(t *testing.T) {
	store := memory.NewStore()
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: "/non/existent/dir",
	}

	// Should return error because ReadDir fails
	err := seeder.Seed(context.Background())
	assert.Error(t, err)
}

func TestSeeder_Seed_EmptyDir(t *testing.T) {
	tempDir := t.TempDir()

	store := memory.NewStore()
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	err := seeder.Seed(context.Background())
	require.NoError(t, err)

	// Should still have hardcoded templates
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, templates)
}
