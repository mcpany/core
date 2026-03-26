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
	// Setup mock store
	store := memory.NewStore()

	// Create temporary directory for examples
	tempDir, err := os.MkdirTemp("", "examples")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a dummy example structure
	// seed.go expects: ExamplesDir / <service_id> / config.yaml
	// Example: examples/foo/config.yaml

	// 1. Valid structure (but parsing is not fully implemented in Seed yet, just logging)
	fooDir := filepath.Join(tempDir, "foo")
	require.NoError(t, os.Mkdir(fooDir, 0755))
	fooConfig := filepath.Join(fooDir, "config.yaml")
	err = os.WriteFile(fooConfig, []byte(`
upstream_services:
- name: foo
  mcp_service:
    stdio_connection:
      command: "echo"
`), 0644)
	require.NoError(t, err)

	// 2. Invalid YAML
	barDir := filepath.Join(tempDir, "bar")
	require.NoError(t, os.Mkdir(barDir, 0755))
	barConfig := filepath.Join(barDir, "config.yaml")
	err = os.WriteFile(barConfig, []byte(`invalid: yaml: :`), 0644)
	require.NoError(t, err)

	// Initialize Seeder
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Run Seed
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify Hardcoded Templates are present
	// We expect "google-calendar", "github", etc.
	expectedTemplates := []string{
		"google-calendar",
		"github",
		"gitlab",
		"slack",
		"notion",
		"linear",
		"jira",
	}

	for _, id := range expectedTemplates {
		tmpl, err := store.GetServiceTemplate(context.Background(), id)
		require.NoError(t, err, "failed to get template %s", id)
		assert.NotNil(t, tmpl, "template %s not found", id)
		assert.Equal(t, id, tmpl.GetId())
	}

	// Verify file scanning logic didn't crash
	// (Since Seed currently only logs errors for file scanning, we can't assert much beyond successful execution)
	// If the logic were implemented to save, we would check for "foo".
	// But currently it doesn't save scanned templates.
	// So we assert that "foo" is NOT present (unless implemented)
	fooTmpl, err := store.GetServiceTemplate(context.Background(), "foo")
	assert.NoError(t, err) // Get returns (nil, nil) if not found in memory store
	assert.Nil(t, fooTmpl, "scanned template 'foo' should not be saved yet as implementation is incomplete")
}
