package config

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaValidationLineNumbers_TypeMismatch(t *testing.T) {
    // address is string, pass int
    yamlContent := `
upstream_services:
  - name: "foo"
    http_service:
      address: 12345
`
    // Line 5: address: 12345

    fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "config.yaml", []byte(yamlContent), 0644))

	store := NewFileStore(fs, []string{"config.yaml"})
	_, err := store.Load(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 5", "Error message should contain line number for type mismatch")
}

func TestSchemaValidationLineNumbers_InvalidDuration(t *testing.T) {
    yamlContent := `
upstream_services:
  - name: "foo"
    command_line_service:
      command: "echo"
      timeout: "bad-duration"
`
    // Line 6: timeout: "bad-duration"

    fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "config.yaml", []byte(yamlContent), 0644))

	store := NewFileStore(fs, []string{"config.yaml"})
	_, err := store.Load(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 6", "Error message should contain line number for invalid duration")
}

func TestSchemaValidationLineNumbers_UnknownField(t *testing.T) {
    yamlContent := `
upstream_services:
  - name: "foo"
    http_service:
       url: "http://localhost:8080"
       unknown_field: "val"
`
    // Line 6: unknown_field: "val"

    fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "config.yaml", []byte(yamlContent), 0644))

	store := NewFileStore(fs, []string{"config.yaml"})
	_, err := store.Load(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 6", "Error message should contain line number for unknown field")
}
