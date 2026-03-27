package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCmd(t *testing.T) {
	// Setup temporary directory and files for testing
	tempDir := t.TempDir()
	validYAMLPath := filepath.Join(tempDir, "valid.yaml")
	invalidYAMLPath := filepath.Join(tempDir, "invalid.yaml")
	validationErrorYAMLPath := filepath.Join(tempDir, "validation_error.yaml")

	validYAMLContent := []byte(`
global:
  log_level: info
upstream_services:
  - id: "test-service"
    http_service:
      address: "http://localhost:8080"
`)
	err := os.WriteFile(validYAMLPath, validYAMLContent, 0644)
	require.NoError(t, err)

	invalidYAMLContent := []byte(`
global:
  log_level: info
upstream_services:
  - id: "test-service"
    http_service:
      address: "http://localhost:8080"
  - invalid_yaml: [
`)
	err = os.WriteFile(invalidYAMLPath, invalidYAMLContent, 0644)
	require.NoError(t, err)

	t.Run("Valid Configuration", func(t *testing.T) {
		viper.Reset()
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"validate", "--config-path", validYAMLPath})
		err := cmd.Execute()
		assert.NoError(t, err)
		out := b.String()
		assert.Contains(t, out, "Configuration successfully validated")
	})

	// 2. Invalid YAML syntax
	validationErrorYAMLContent := []byte(`
global:
  log_level: info
upstream_services:
  - id: "test-service"
    http_service:
      address: "invalid-url" # Missing scheme
`)
	err = os.WriteFile(validationErrorYAMLPath, validationErrorYAMLContent, 0644)
	require.NoError(t, err)

	t.Run("Invalid YAML Syntax", func(t *testing.T) {
		viper.Reset()
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"validate", "--config-path", invalidYAMLPath})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal")
	})

	// 3. Validation Error (Invalid HTTP address)
	t.Run("Missing Config File", func(t *testing.T) {
		viper.Reset()
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"validate", "--config-path", "non_existent_file.yaml"})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
	})

	// 4. Default Config Path
	defaultConfigPath := filepath.Join(".", "mcpany.yaml")
	// Make sure we are creating/deleting the correct default
	err = os.WriteFile(defaultConfigPath, validYAMLContent, 0644)
	require.NoError(t, err)
	defer os.Remove(defaultConfigPath)

	t.Run("Validation Error", func(t *testing.T) {
		viper.Reset()
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"validate", "--config-path", validationErrorYAMLPath})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Default Config Path", func(t *testing.T) {
		viper.Reset()
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"validate"}) // should pick up ./mcpany.yaml
		err := cmd.Execute()
		assert.NoError(t, err)
		out := b.String()
		assert.Contains(t, out, "Configuration successfully validated")
	})

}
