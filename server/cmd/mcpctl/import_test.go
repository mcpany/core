// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestImportCmd(t *testing.T) {
	tempDir := t.TempDir()

	validJSON := `{
  "mcpServers": {
    "test-server": {
      "command": "node",
      "args": ["server.js"],
      "env": {
        "API_KEY": "12345"
      }
    }
  }
}`
	validJSONPath := filepath.Join(tempDir, "config.json")
	err := os.WriteFile(validJSONPath, []byte(validJSON), 0644)
	require.NoError(t, err)

	invalidJSONPath := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidJSONPath, []byte("{ invalid json"), 0644)
	require.NoError(t, err)

	emptyJSONPath := filepath.Join(tempDir, "empty.json")
	err = os.WriteFile(emptyJSONPath, []byte(`{"mcpServers": {}}`), 0644)
	require.NoError(t, err)

	t.Run("Happy Path (Stdout)", func(t *testing.T) {
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"import", validJSONPath})
		err := cmd.Execute()
		assert.NoError(t, err)

		output := b.String()
		assert.Contains(t, output, "upstream_services:")
		assert.Contains(t, output, "name: test-server")
		assert.Contains(t, output, "command: node")
		assert.Contains(t, output, "args:")
		assert.Contains(t, output, "- server.js")
		assert.Contains(t, output, "env:")
		assert.Contains(t, output, "API_KEY: \"12345\"")
	})

	t.Run("Happy Path (File Output)", func(t *testing.T) {
		outputPath := filepath.Join(tempDir, "output.yaml")
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"import", validJSONPath, "-o", outputPath})
		err := cmd.Execute()
		assert.NoError(t, err)

		assert.Contains(t, b.String(), "Successfully imported configuration")

		// Verify file content
		content, err := os.ReadFile(outputPath)
		require.NoError(t, err)

		var config McpAnyConfig
		err = yaml.Unmarshal(content, &config)
		require.NoError(t, err)

		require.Len(t, config.UpstreamServices, 1)
		service := config.UpstreamServices[0]
		assert.Equal(t, "test-server", service.Name)
		require.NotNil(t, service.McpService)
		require.NotNil(t, service.McpService.StdioConnection)
		assert.Equal(t, "node", service.McpService.StdioConnection.Command)
		assert.Equal(t, []string{"server.js"}, service.McpService.StdioConnection.Args)
		assert.Equal(t, map[string]string{"API_KEY": "12345"}, service.McpService.StdioConnection.Env)
	})

	t.Run("File Not Found", func(t *testing.T) {
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"import", filepath.Join(tempDir, "non-existent.json")})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read input file")
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"import", invalidJSONPath})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse Claude Desktop config")
	})

	t.Run("Empty Config", func(t *testing.T) {
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"import", emptyJSONPath})
		err := cmd.Execute()
		assert.NoError(t, err)

		output := b.String()
		assert.Contains(t, output, "upstream_services: []")
		// Or YAML might represent empty list as nothing or just the key.
		// Let's parse it to be sure.
		var config McpAnyConfig
		err = yaml.Unmarshal([]byte(output), &config)
		require.NoError(t, err)
		assert.Empty(t, config.UpstreamServices)
	})

	t.Run("Complex Config", func(t *testing.T) {
		complexJSON := `{
  "mcpServers": {
    "complex": {
      "command": "python",
      "args": ["-m", "server", "--port", "8080"],
      "env": {
        "ENV_VAR_1": "value1",
        "ENV_VAR_2": "value 2"
      }
    }
  }
}`
		complexPath := filepath.Join(tempDir, "complex.json")
		err := os.WriteFile(complexPath, []byte(complexJSON), 0644)
		require.NoError(t, err)

		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"import", complexPath})
		err = cmd.Execute()
		assert.NoError(t, err)

		output := b.String()
		// Verify structure roughly, but rely on parsing for exactness
		assert.Contains(t, output, "complex")
		assert.Contains(t, output, "python")

		var config McpAnyConfig
		err = yaml.Unmarshal([]byte(output), &config)
		require.NoError(t, err)

		require.Len(t, config.UpstreamServices, 1)
		service := config.UpstreamServices[0]
		conn := service.McpService.StdioConnection
		assert.Equal(t, "python", conn.Command)
		assert.Equal(t, []string{"-m", "server", "--port", "8080"}, conn.Args)
		assert.Equal(t, "value1", conn.Env["ENV_VAR_1"])
		assert.Equal(t, "value 2", conn.Env["ENV_VAR_2"])
	})

    t.Run("Missing Argument", func(t *testing.T) {
        cmd := newRootCmd()
        b := bytes.NewBufferString("")
        cmd.SetOut(b)
        cmd.SetArgs([]string{"import"})
        err := cmd.Execute()
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
    })

    t.Run("Invalid Output Path", func(t *testing.T) {
        // Create a directory where the file should be to trigger error (or use an invalid path like /root/foo)
        // Using a directory as file path usually fails.
        badPath := filepath.Join(tempDir, "bad_dir")
        err := os.Mkdir(badPath, 0755)
        require.NoError(t, err)

        cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"import", validJSONPath, "-o", badPath})
		err = cmd.Execute()
		assert.Error(t, err)
        // Error message depends on OS, but usually "is a directory" or "access denied"
        // Just checking error exists is enough for coverage.
    })
}
