package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCmd(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Valid Configuration
	validConfig := `
global_settings:
  mcp_listen_address: "localhost:50050"
upstream_services:
  - name: "my-service"
    http_service:
      address: "http://example.com"
      tools:
        - name: "my-tool"
          call_id: "my-call"
      calls:
        my-call:
          endpoint_path: "/api"
          method: "HTTP_METHOD_GET"
`
	validConfigPath := filepath.Join(tempDir, "valid_config.yaml")
	err := os.WriteFile(validConfigPath, []byte(validConfig), 0644)
	require.NoError(t, err)

	t.Run("Valid Configuration", func(t *testing.T) {
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"validate", "--config-path", validConfigPath})
		err := cmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, b.String(), "Configuration is valid.")
	})

	// 2. Invalid YAML Syntax
	invalidYAML := `
global_settings:
  mcp_listen_address: "localhost:50050"
upstream_services:
  - name: "my-service"
    http_service:
      address: "http://example.com"
      tools:
        - name: "my-tool"
          call_id: "my-call"
      calls:
        my-call:
          endpoint_path: "/api"
          method: "HTTP_METHOD_GET"
    invalid_indentation
`
	invalidYAMLPath := filepath.Join(tempDir, "invalid_yaml.yaml")
	err = os.WriteFile(invalidYAMLPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	t.Run("Invalid YAML Syntax", func(t *testing.T) {
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"validate", "--config-path", invalidYAMLPath})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal config")
	})

	// 3. Validation Error (Invalid HTTP address)
	invalidConfig := `
global_settings:
  mcp_listen_address: "localhost:50050"
upstream_services:
  - name: "my-service"
    http_service:
      address: "::invalid::"
      tools:
        - name: "my-tool"
          call_id: "my-call"
      calls:
        my-call:
          endpoint_path: "/api"
          method: "HTTP_METHOD_GET"
`
	invalidConfigPath := filepath.Join(tempDir, "invalid_config.yaml")
	err = os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0644)
	require.NoError(t, err)

	t.Run("Validation Error", func(t *testing.T) {
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"validate", "--config-path", invalidConfigPath})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Configuration Validation Failed")
		assert.Contains(t, err.Error(), "invalid http address")
	})
}
