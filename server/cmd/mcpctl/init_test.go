// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitCmd_HTTP(t *testing.T) {
	cmd := newInitCmd()

	// Simulate user input
	input := "test-http-api\n1\nhttp://example.com/api\n"
	cmd.SetIn(bytes.NewBufferString(input))

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "mcp.yaml")
	cmd.SetArgs([]string{"--output", outPath})

	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outPath)
	assert.NoError(t, err)
	yamlStr := string(content)

	assert.Contains(t, yamlStr, "name: test-http-api")
	assert.Contains(t, yamlStr, "http_service:")
	assert.Contains(t, yamlStr, "address: http://example.com/api")
}

func TestInitCmd_GRPC(t *testing.T) {
	cmd := newInitCmd()

	// Simulate user input
	input := "test-grpc-api\n2\nlocalhost:50051\n"
	cmd.SetIn(bytes.NewBufferString(input))

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "mcp.yaml")
	cmd.SetArgs([]string{"--output", outPath})

	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outPath)
	assert.NoError(t, err)
	yamlStr := string(content)

	assert.Contains(t, yamlStr, "name: test-grpc-api")
	assert.Contains(t, yamlStr, "grpc_service:")
	assert.Contains(t, yamlStr, "address: localhost:50051")
	assert.Contains(t, yamlStr, "use_reflection: true")
}

func TestInitCmd_Command(t *testing.T) {
	cmd := newInitCmd()

	// Simulate user input
	input := "test-cmd-api\n3\npython3\napp.py arg1\n"
	cmd.SetIn(bytes.NewBufferString(input))

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "mcp.yaml")
	cmd.SetArgs([]string{"--output", outPath})

	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outPath)
	assert.NoError(t, err)
	yamlStr := string(content)

	assert.Contains(t, yamlStr, "name: test-cmd-api")
	assert.Contains(t, yamlStr, "mcp_service:")
	assert.Contains(t, yamlStr, "stdio_connection:")
	assert.Contains(t, yamlStr, "command: python3")
	assert.Contains(t, yamlStr, "- app.py")
	assert.Contains(t, yamlStr, "- arg1")
}

func TestInitCmd_Defaults(t *testing.T) {
	cmd := newInitCmd()

	// Simulate empty inputs
	input := "\n\n\n"
	cmd.SetIn(bytes.NewBufferString(input))

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "mcp.yaml")
	cmd.SetArgs([]string{"--output", outPath})

	err := cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outPath)
	assert.NoError(t, err)
	yamlStr := string(content)

	// Defaults to my-service and HTTP type (1)
	assert.Contains(t, yamlStr, "name: my-service")
	assert.Contains(t, yamlStr, "http_service:")
}
