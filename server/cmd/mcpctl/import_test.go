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
)

func TestImportCmd_Stdout(t *testing.T) {
	// 1. Create temporary input file
	inputContent := `{
  "mcpServers": {
    "test-server": {
      "command": "echo",
      "args": ["hello", "world"],
      "env": {"TEST_ENV": "value"}
    }
  }
}`
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "config.json")
	err := os.WriteFile(inputPath, []byte(inputContent), 0600)
	require.NoError(t, err)

	// 2. Execute command
	cmd := newImportCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{inputPath})
	err = cmd.Execute()

	// 3. Verify
	assert.NoError(t, err)
	output := b.String()
	assert.Contains(t, output, "upstream_services:")
	assert.Contains(t, output, "- name: test-server")
	assert.Contains(t, output, "command: echo")
	assert.Contains(t, output, "- hello")
	assert.Contains(t, output, "- world")
	assert.Contains(t, output, "TEST_ENV: value")
}

func TestImportCmd_FileOutput(t *testing.T) {
	// 1. Create temporary input file
	inputContent := `{
  "mcpServers": {
    "file-server": {
      "command": "cat",
      "args": ["file.txt"]
    }
  }
}`
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "config.json")
	outputPath := filepath.Join(tmpDir, "output.yaml")
	err := os.WriteFile(inputPath, []byte(inputContent), 0600)
	require.NoError(t, err)

	// 2. Execute command
	cmd := newImportCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{inputPath, "--output", outputPath})
	err = cmd.Execute()

	// 3. Verify execution
	assert.NoError(t, err)
	assert.Contains(t, b.String(), "Successfully imported configuration to")
	assert.Contains(t, b.String(), outputPath)

	// 4. Verify file content
	outputBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	output := string(outputBytes)
	assert.Contains(t, output, "upstream_services:")
	assert.Contains(t, output, "- name: file-server")
	assert.Contains(t, output, "command: cat")
}

func TestImportCmd_FileNotFound(t *testing.T) {
	cmd := newImportCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"/non/existent/path/config.json"})
	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read input file")
}

func TestImportCmd_InvalidJSON(t *testing.T) {
	// 1. Create temporary invalid input file
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "invalid.json")
	err := os.WriteFile(inputPath, []byte("{invalid-json"), 0600)
	require.NoError(t, err)

	// 2. Execute command
	cmd := newImportCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{inputPath})
	err = cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse Claude Desktop config")
}

func TestImportCmd_EmptyConfig(t *testing.T) {
	// 1. Create temporary empty config file
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "empty.json")
	err := os.WriteFile(inputPath, []byte(`{"mcpServers": {}}`), 0600)
	require.NoError(t, err)

	// 2. Execute command
	cmd := newImportCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{inputPath})
	err = cmd.Execute()

	// 3. Verify
	assert.NoError(t, err)
	output := b.String()
	assert.Contains(t, output, "upstream_services: []")
}
