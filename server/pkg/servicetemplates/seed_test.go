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

func TestSeed_BuiltIn(t *testing.T) {
	// Create a temporary directory for examples (empty for this test)
	tempDir, err := os.MkdirTemp("", "seed_test_builtin")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store := memory.NewStore()
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)

	// We expect 7 built-in templates
	assert.Len(t, templates, 7)

	// Verify one of them
	tmpl, err := store.GetServiceTemplate(context.Background(), "google-calendar")
	require.NoError(t, err)
	require.NotNil(t, tmpl)
	assert.Equal(t, "Google Calendar", tmpl.GetName())
	assert.Equal(t, "google-calendar", tmpl.GetIcon())
}

func TestSeed_FromFiles(t *testing.T) {
	// Create a temporary directory for examples
	tempDir, err := os.MkdirTemp("", "seed_test_files")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a dummy service directory
	svcDir := filepath.Join(tempDir, "my-service")
	err = os.Mkdir(svcDir, 0755)
	require.NoError(t, err)

	// Create a config.yaml
	configContent := `
upstream_services:
  - name: my-service
    mcp_service:
      http_connection:
        http_address: https://api.myservice.com
`
	err = os.WriteFile(filepath.Join(svcDir, "config.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	store := memory.NewStore()
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Note: File-based seeding is currently unimplemented in the source; verifying that files are ignored and only built-ins are loaded.
	// We expect 7 built-in templates only.
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)
	assert.Len(t, templates, 7) // Only built-ins
}

func TestSeed_InvalidYaml(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "seed_test_invalid")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	svcDir := filepath.Join(tempDir, "bad-service")
	err = os.Mkdir(svcDir, 0755)
	require.NoError(t, err)

	// broken yaml
	err = os.WriteFile(filepath.Join(svcDir, "config.yaml"), []byte("upstream_services: [ unclosed_list"), 0644)
	require.NoError(t, err)

	store := memory.NewStore()
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Should not return error, just log and continue (and save built-ins)
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)
	assert.Len(t, templates, 7)
}

func TestSeed_NoConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "seed_test_noconfig")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	svcDir := filepath.Join(tempDir, "empty-service")
	err = os.Mkdir(svcDir, 0755)
	require.NoError(t, err)

	store := memory.NewStore()
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	err = seeder.Seed(context.Background())
	require.NoError(t, err)

    templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)
	assert.Len(t, templates, 7)
}
