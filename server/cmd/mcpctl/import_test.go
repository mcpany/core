// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestImportCmd_HappyPath_Stdout(t *testing.T) {
	// Create a temporary input file
	inputContent := `{
		"mcpServers": {
			"test-server": {
				"command": "python",
				"args": ["server.py"],
				"env": {
					"API_KEY": "12345"
				}
			}
		}
	}`
	inputFile, err := os.CreateTemp("", "claude_config_*.json")
	require.NoError(t, err)
	defer os.Remove(inputFile.Name())
	_, err = inputFile.WriteString(inputContent)
	require.NoError(t, err)
	inputFile.Close()

	// Capture stdout
	var out bytes.Buffer

	cmd := newImportCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{inputFile.Name()})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output
	var outputConfig McpAnyConfig
	err = yaml.Unmarshal(out.Bytes(), &outputConfig)
	require.NoError(t, err)

	require.Len(t, outputConfig.UpstreamServices, 1)
	svc := outputConfig.UpstreamServices[0]
	assert.Equal(t, "test-server", svc.Name)
	assert.NotNil(t, svc.McpService)
	assert.NotNil(t, svc.McpService.StdioConnection)
	assert.Equal(t, "python", svc.McpService.StdioConnection.Command)
	assert.Equal(t, []string{"server.py"}, svc.McpService.StdioConnection.Args)
	assert.Equal(t, map[string]string{"API_KEY": "12345"}, svc.McpService.StdioConnection.Env)
}

func TestImportCmd_HappyPath_File(t *testing.T) {
	inputContent := `{
		"mcpServers": {
			"server1": {
				"command": "node",
				"args": ["index.js"]
			}
		}
	}`
	inputFile, err := os.CreateTemp("", "claude_config_*.json")
	require.NoError(t, err)
	defer os.Remove(inputFile.Name())
	_, err = inputFile.WriteString(inputContent)
	require.NoError(t, err)
	inputFile.Close()

	outputFile, err := os.CreateTemp("", "mcp_config_*.yaml")
	require.NoError(t, err)
	outputFilePath := outputFile.Name()
	outputFile.Close() // Close it so the command can write to it
	defer os.Remove(outputFilePath)

	// Execute command with -o flag
	cmd := newImportCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{inputFile.Name(), "-o", outputFilePath})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify success message
	assert.Contains(t, out.String(), "Successfully imported configuration to")

	// Verify file content
	outputContent, err := os.ReadFile(outputFilePath)
	require.NoError(t, err)

	var outputConfig McpAnyConfig
	err = yaml.Unmarshal(outputContent, &outputConfig)
	require.NoError(t, err)

	require.Len(t, outputConfig.UpstreamServices, 1)
	assert.Equal(t, "server1", outputConfig.UpstreamServices[0].Name)
}

func TestImportCmd_FileNotFound(t *testing.T) {
	cmd := newImportCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"/non/existent/path/config.json"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read input file")
}

func TestImportCmd_InvalidJSON(t *testing.T) {
	inputFile, err := os.CreateTemp("", "invalid_config_*.json")
	require.NoError(t, err)
	defer os.Remove(inputFile.Name())
	_, err = inputFile.WriteString("{ invalid json }")
	require.NoError(t, err)
	inputFile.Close()

	cmd := newImportCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{inputFile.Name()})

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse Claude Desktop config")
}

func TestImportCmd_EmptyConfig(t *testing.T) {
	inputContent := `{ "mcpServers": {} }`
	inputFile, err := os.CreateTemp("", "empty_config_*.json")
	require.NoError(t, err)
	defer os.Remove(inputFile.Name())
	_, err = inputFile.WriteString(inputContent)
	require.NoError(t, err)
	inputFile.Close()

	var out bytes.Buffer
	cmd := newImportCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{inputFile.Name()})

	err = cmd.Execute()
	require.NoError(t, err)

	var outputConfig McpAnyConfig
	err = yaml.Unmarshal(out.Bytes(), &outputConfig)
	require.NoError(t, err)
	assert.Empty(t, outputConfig.UpstreamServices)
}

func TestImportCmd_ComplexConfig(t *testing.T) {
	inputContent := `{
		"mcpServers": {
			"complex-server": {
				"command": "/usr/bin/env",
				"args": ["python", "-m", "server", "--port", "8080"],
				"env": {
					"DEBUG": "true",
					"LOG_LEVEL": "info"
				}
			},
			"simple-server": {
				"command": "echo",
				"args": ["hello"]
			}
		}
	}`
	inputFile, err := os.CreateTemp("", "complex_config_*.json")
	require.NoError(t, err)
	defer os.Remove(inputFile.Name())
	_, err = inputFile.WriteString(inputContent)
	require.NoError(t, err)
	inputFile.Close()

	var out bytes.Buffer
	cmd := newImportCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{inputFile.Name()})

	err = cmd.Execute()
	require.NoError(t, err)

	var outputConfig McpAnyConfig
	err = yaml.Unmarshal(out.Bytes(), &outputConfig)
	require.NoError(t, err)

	require.Len(t, outputConfig.UpstreamServices, 2)

	// Since map iteration order is random, we need to find the specific services
	var complexSvc, simpleSvc UpstreamService
	for _, s := range outputConfig.UpstreamServices {
		if s.Name == "complex-server" {
			complexSvc = s
		} else if s.Name == "simple-server" {
			simpleSvc = s
		}
	}

	assert.Equal(t, "complex-server", complexSvc.Name)
	assert.Equal(t, "/usr/bin/env", complexSvc.McpService.StdioConnection.Command)
	assert.Equal(t, []string{"python", "-m", "server", "--port", "8080"}, complexSvc.McpService.StdioConnection.Args)
	assert.Equal(t, "true", complexSvc.McpService.StdioConnection.Env["DEBUG"])

	assert.Equal(t, "simple-server", simpleSvc.Name)
	assert.Equal(t, "echo", simpleSvc.McpService.StdioConnection.Command)
	assert.Nil(t, simpleSvc.McpService.StdioConnection.Env) // Should be nil or empty
}

func TestImportCmd_InvalidArgs(t *testing.T) {
    cmd := newImportCmd()
    var out bytes.Buffer
    cmd.SetOut(&out)
    // No args provided
    cmd.SetArgs([]string{})

    err := cmd.Execute()
    require.Error(t, err)
    // Cobra usually returns an error about missing arguments
}

func TestImportCmd_BadOutputPath(t *testing.T) {
    // This test might be tricky depending on permissions, but trying to write to a directory usually fails
    inputContent := `{"mcpServers": {}}`
	inputFile, err := os.CreateTemp("", "valid_config_*.json")
	require.NoError(t, err)
	defer os.Remove(inputFile.Name())
	_, err = inputFile.WriteString(inputContent)
	require.NoError(t, err)
	inputFile.Close()

    // Create a temp dir to use as a file path which should fail
    tempDir, err := os.MkdirTemp("", "test_dir")
    require.NoError(t, err)
    defer os.RemoveAll(tempDir)

    cmd := newImportCmd()
    var out bytes.Buffer
    cmd.SetOut(&out)
    cmd.SetArgs([]string{inputFile.Name(), "-o", tempDir}) // Writing to a directory path usually fails

    err = cmd.Execute()
    require.Error(t, err)
    // The error message depends on OS, but usually "is a directory"
}
