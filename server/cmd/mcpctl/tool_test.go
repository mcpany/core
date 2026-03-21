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

func TestToolHashCmd(t *testing.T) {
	tempDir := t.TempDir()

	configContent := `
global_settings:
  mcp_listen_address: "localhost:50050"
upstream_services:
  - name: "http-service"
    http_service:
      address: "http://example.com"
      tools:
        - name: "http-tool"
          call_id: "http-call"
      calls:
        http-call:
          endpoint_path: "/api"
          method: "HTTP_METHOD_GET"
  - name: "grpc-service"
    grpc_service:
      address: "localhost:50051"
      tools:
        - name: "grpc-tool"
          call_id: "grpc-call"
      calls:
        grpc-call:
          method: "MyService/MyMethod"
`
	configPath := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	t.Run("Existing Tool (HTTP)", func(t *testing.T) {
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"tool", "hash", "http-tool", "--config-path", configPath})
		err := cmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, b.String(), "Tool: http-tool")
		assert.Contains(t, b.String(), "Hash:")
	})

	t.Run("Existing Tool (gRPC)", func(t *testing.T) {
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"tool", "hash", "grpc-tool", "--config-path", configPath})
		err := cmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, b.String(), "Tool: grpc-tool")
		assert.Contains(t, b.String(), "Hash:")
	})

	t.Run("Non-Existent Tool", func(t *testing.T) {
		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"tool", "hash", "non-existent-tool", "--config-path", configPath})
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tool \"non-existent-tool\" not found")
	})

    t.Run("Invalid Configuration", func(t *testing.T) {
        cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
        // Point to a non-existent config file to trigger load error
		cmd.SetArgs([]string{"tool", "hash", "http-tool", "--config-path", filepath.Join(tempDir, "non-existent.yaml")})
		err := cmd.Execute()
		assert.Error(t, err)
        assert.Contains(t, err.Error(), "failed to load configuration")
    })
}
