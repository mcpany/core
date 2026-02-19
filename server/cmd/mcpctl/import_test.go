// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestImportCmd(t *testing.T) {
	// Helper to create a temporary Claude config file
	createTempClaudeConfig := func(t *testing.T, config ClaudeDesktopConfig) string {
		t.Helper()
		data, err := json.Marshal(config)
		require.NoError(t, err)

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "claude_desktop_config.json")
		err = os.WriteFile(tmpFile, data, 0600)
		require.NoError(t, err)
		return tmpFile
	}

	t.Run("Happy Path - Stdout", func(t *testing.T) {
		inputConfig := ClaudeDesktopConfig{
			MCPServers: map[string]MCPServerConfig{
				"test-server": {
					Command: "docker",
					Args:    []string{"run", "-i", "--rm", "test-image"},
					Env:     map[string]string{"DEBUG": "true"},
				},
			},
		}
		inputFile := createTempClaudeConfig(t, inputConfig)

		cmd := newImportCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{inputFile})

		err := cmd.Execute()
		assert.NoError(t, err)

		var outputConfig McpAnyConfig
		err = yaml.Unmarshal(b.Bytes(), &outputConfig)
		assert.NoError(t, err)
		assert.Len(t, outputConfig.UpstreamServices, 1)

		svc := outputConfig.UpstreamServices[0]
		assert.Equal(t, "test-server", svc.Name)
		assert.NotNil(t, svc.McpService)
		assert.NotNil(t, svc.McpService.StdioConnection)
		assert.Equal(t, "docker", svc.McpService.StdioConnection.Command)
		assert.Equal(t, []string{"run", "-i", "--rm", "test-image"}, svc.McpService.StdioConnection.Args)
		assert.Equal(t, map[string]string{"DEBUG": "true"}, svc.McpService.StdioConnection.Env)
	})

	t.Run("Happy Path - File Output", func(t *testing.T) {
		inputConfig := ClaudeDesktopConfig{
			MCPServers: map[string]MCPServerConfig{
				"file-server": {
					Command: "node",
					Args:    []string{"index.js"},
				},
			},
		}
		inputFile := createTempClaudeConfig(t, inputConfig)

		tmpDir := t.TempDir()
		outputFile := filepath.Join(tmpDir, "output.yaml")

		cmd := newImportCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{inputFile, "--output", outputFile})

		err := cmd.Execute()
		assert.NoError(t, err)

		// Verify file created
		data, err := os.ReadFile(outputFile)
		assert.NoError(t, err)

		var outputConfig McpAnyConfig
		err = yaml.Unmarshal(data, &outputConfig)
		assert.NoError(t, err)
		assert.Len(t, outputConfig.UpstreamServices, 1)
		assert.Equal(t, "file-server", outputConfig.UpstreamServices[0].Name)
	})

	t.Run("File Not Found", func(t *testing.T) {
		cmd := newImportCmd()
		cmd.SetArgs([]string{"/non/existent/path/config.json"})

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read input file")
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(inputFile, []byte("{ invalid json"), 0600)
		require.NoError(t, err)

		cmd := newImportCmd()
		cmd.SetArgs([]string{inputFile})

		err = cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse Claude Desktop config")
	})

	t.Run("Empty Config", func(t *testing.T) {
		inputConfig := ClaudeDesktopConfig{
			MCPServers: map[string]MCPServerConfig{},
		}
		inputFile := createTempClaudeConfig(t, inputConfig)

		cmd := newImportCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{inputFile})

		err := cmd.Execute()
		assert.NoError(t, err)

		var outputConfig McpAnyConfig
		err = yaml.Unmarshal(b.Bytes(), &outputConfig)
		assert.NoError(t, err)
		assert.Len(t, outputConfig.UpstreamServices, 0)
	})

	t.Run("Complex Config", func(t *testing.T) {
		inputConfig := ClaudeDesktopConfig{
			MCPServers: map[string]MCPServerConfig{
				"complex": {
					Command: "cmd",
					Args:    []string{"arg1", "arg2"},
					Env:     map[string]string{"KEY1": "VAL1", "KEY2": "VAL2"},
				},
				"another": {
					Command: "other",
				},
			},
		}
		inputFile := createTempClaudeConfig(t, inputConfig)

		cmd := newImportCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{inputFile})

		err := cmd.Execute()
		assert.NoError(t, err)

		var outputConfig McpAnyConfig
		err = yaml.Unmarshal(b.Bytes(), &outputConfig)
		assert.NoError(t, err)
		assert.Len(t, outputConfig.UpstreamServices, 2)

		// Map order is undefined in Go iteration, so find by name
		for _, svc := range outputConfig.UpstreamServices {
			if svc.Name == "complex" {
				assert.Equal(t, "cmd", svc.McpService.StdioConnection.Command)
				assert.Len(t, svc.McpService.StdioConnection.Args, 2)
				assert.Equal(t, "VAL1", svc.McpService.StdioConnection.Env["KEY1"])
			} else if svc.Name == "another" {
				assert.Equal(t, "other", svc.McpService.StdioConnection.Command)
			} else {
				t.Errorf("Unexpected service name: %s", svc.Name)
			}
		}
	})
}
