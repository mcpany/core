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

	// Create valid example directory
	validDir := filepath.Join(tempDir, "valid-service")
	require.NoError(t, os.Mkdir(validDir, 0755))

	validConfig := `
upstream_services:
  - name: valid-service
    mcp_service:
      http_connection:
        http_address: "https://api.example.com"
`
	require.NoError(t, os.WriteFile(filepath.Join(validDir, "config.yaml"), []byte(validConfig), 0644))

	// Create invalid example directory (invalid yaml)
	invalidDir := filepath.Join(tempDir, "invalid-service")
	require.NoError(t, os.Mkdir(invalidDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(invalidDir, "config.yaml"), []byte("invalid: yaml: :"), 0644))

	// Create empty example directory (no config.yaml)
	emptyDir := filepath.Join(tempDir, "empty-service")
	require.NoError(t, os.Mkdir(emptyDir, 0755))

	// Create store
	store := memory.NewStore()

	seeder := &Seeder{
		Store:       store,
		ExamplesDir: tempDir,
	}

	// Run Seed
	err = seeder.Seed(context.Background())
	require.NoError(t, err)

	// Verify hardcoded templates are present
	templates, err := store.ListServiceTemplates(context.Background())
	require.NoError(t, err)

	expectedTemplates := []string{
		"google-calendar",
		"github",
		"gitlab",
		"slack",
		"notion",
		"linear",
		"jira",
	}

	for _, expectedID := range expectedTemplates {
		found := false
		for _, t := range templates {
			if t.GetId() == expectedID {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected template %s to be seeded", expectedID)
	}

	// Verify that the file scanning logic didn't crash
	// Currently, the file scanning logic doesn't save anything to the store,
	// so we can't assert that "valid-service" is present.
	// But running Seed successfully proves that it handled the files gracefully.

	// Ensure that "valid-service" is NOT in the templates (since the logic is incomplete/no-op)
	// This confirms our understanding of the current behavior.
	foundValid := false
	for _, t := range templates {
		if t.GetId() == "valid-service" {
			foundValid = true
			break
		}
	}
	assert.False(t, foundValid, "Did not expect valid-service to be seeded (logic is incomplete)")
}
