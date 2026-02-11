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

func TestSeeder_Seed_Hardcoded(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()

	// Create a temporary directory for examples to avoid error
	tempDir, err := os.MkdirTemp("", "seed_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	err = seeder.Seed(ctx)
	require.NoError(t, err)

	// Verify hardcoded templates
	templates, err := store.ListServiceTemplates(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, templates)

	// Check specifically for "github"
	foundGithub := false
	for _, tmpl := range templates {
		if tmpl.GetId() == "github" {
			foundGithub = true
			break
		}
	}
	assert.True(t, foundGithub, "github template should be seeded")

	// Check for "google-calendar"
	foundCalendar := false
	for _, tmpl := range templates {
		if tmpl.GetId() == "google-calendar" {
			foundCalendar = true
			break
		}
	}
	assert.True(t, foundCalendar, "google-calendar template should be seeded")
}

func TestSeeder_Seed_FileScanning(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()

	tempDir, err := os.MkdirTemp("", "seed_scan_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a valid config structure
	// service_a/config.yaml
	serviceDir := filepath.Join(tempDir, "service_a")
	err = os.Mkdir(serviceDir, 0755)
	require.NoError(t, err)

	configContent := `
upstream_services:
- name: service_a
  upstream_auth:
    bearer_token:
      token:
        plain_text: "token"
`
	err = os.WriteFile(filepath.Join(serviceDir, "config.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Should not crash
	err = seeder.Seed(ctx)
	require.NoError(t, err)
}

func TestSeeder_Seed_InvalidYAML(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()

	tempDir, err := os.MkdirTemp("", "seed_invalid_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create an invalid config
	serviceDir := filepath.Join(tempDir, "service_b")
	err = os.Mkdir(serviceDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(serviceDir, "config.yaml"), []byte("invalid: yaml: :"), 0644)
	require.NoError(t, err)

	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Should not return error, just log to stdout and continue
	err = seeder.Seed(ctx)
	require.NoError(t, err)
}

func TestSeeder_Seed_ReadDirError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()

	// Point to a non-existent directory
	seeder := &Seeder{
		Store:       store,
		ExamplesDir: "/path/to/non/existent/dir",
	}

	err := seeder.Seed(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read examples dir")
}
