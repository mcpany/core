package config

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestMissingEnvVar(t *testing.T) {
	// Create a mock filesystem
	fs := afero.NewMemMapFs()
	configContent := `
upstream_services:
  - name: "my-service"
    http_service:
      address: "http://localhost:${PORT}"
`
	afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)

	// Ensure PORT is not set
	t.Setenv("PORT", "")

	store := NewFileStore(fs, []string{"config.yaml"})
	_, err := store.Load(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Line 5: variable ${PORT} is missing")
	assert.Contains(t, err.Error(), "Fix: Set these environment variables")
}
