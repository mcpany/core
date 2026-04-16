// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitCmd(t *testing.T) {
	// Create a temporary directory to avoid overwriting existing files
	tempDir, err := os.MkdirTemp("", "mcpctl-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	outFile := tempDir + "/mcp.yaml"

	cmd := newRootCmd()

	// Capture output
	b := bytes.NewBufferString("")
	cmd.SetOut(b)

	// Simulate user input for choice 1 (HTTP)
	inBuf := bytes.NewBufferString("test-service\n1\nhttp://example.com\n")
	cmd.SetIn(inBuf)

	cmd.SetArgs([]string{"init", "--output", outFile})
	err = cmd.Execute()
	assert.NoError(t, err)

	// Verify output
	assert.Contains(t, b.String(), "Success! Generated configuration saved to")

	// Read generated file
	content, err := os.ReadFile(outFile)
	assert.NoError(t, err)
	yamlStr := string(content)

	assert.Contains(t, yamlStr, "name: test-service")
	assert.Contains(t, yamlStr, "http_service:")
	assert.Contains(t, yamlStr, "address: http://example.com")
}

func TestInitCmd_GRPC(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpctl-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	outFile := tempDir + "/mcp.yaml"

	cmd := newRootCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)

	// Simulate user input for choice 2 (gRPC)
	inBuf := bytes.NewBufferString("test-grpc\n2\nlocalhost:9090\n")
	cmd.SetIn(inBuf)

	cmd.SetArgs([]string{"init", "--output", outFile})
	err = cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outFile)
	assert.NoError(t, err)
	yamlStr := string(content)

	assert.Contains(t, yamlStr, "name: test-grpc")
	assert.Contains(t, yamlStr, "grpc_service:")
	assert.Contains(t, yamlStr, "address: localhost:9090")
	assert.Contains(t, yamlStr, "use_reflection: true")
}

func TestInitCmd_Stdio(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpctl-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	outFile := tempDir + "/mcp.yaml"

	cmd := newRootCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)

	// Simulate user input for choice 3 (Stdio)
	inBuf := bytes.NewBufferString("test-stdio\n3\npython3\napp.py arg1\n")
	cmd.SetIn(inBuf)

	cmd.SetArgs([]string{"init", "--output", outFile})
	err = cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outFile)
	assert.NoError(t, err)
	yamlStr := string(content)

	assert.Contains(t, yamlStr, "name: test-stdio")
	assert.Contains(t, yamlStr, "mcp_service:")
	assert.Contains(t, yamlStr, "stdio_connection:")
	assert.Contains(t, yamlStr, "command: python3")
	assert.Contains(t, yamlStr, "- app.py")
	assert.Contains(t, yamlStr, "- arg1")
}

func TestInitCmd_Defaults(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpctl-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	outFile := tempDir + "/mcp.yaml"

	cmd := newRootCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)

	// Simulate empty inputs
	inBuf := bytes.NewBufferString("\n\n\n")
	cmd.SetIn(inBuf)

	cmd.SetArgs([]string{"init", "--output", outFile})
	err = cmd.Execute()
	assert.NoError(t, err)

	content, err := os.ReadFile(outFile)
	assert.NoError(t, err)
	yamlStr := string(content)

	assert.Contains(t, yamlStr, "name: my-service")
	assert.Contains(t, yamlStr, "http_service:")
}
