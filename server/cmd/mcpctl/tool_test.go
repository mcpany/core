// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolHashCmd(t *testing.T) {
	configContent := `
upstream_services:
  - name: "my-service"
    http_service:
      address: "http://example.com"
      tools:
        - name: "my-tool"
          description: "A test tool"
          input_schema:
            type: "object"
            properties:
              query: { type: "string" }
`
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := newRootCmd(fs)
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"tool", "hash", "my-tool", "--config-path", "/config.yaml"})

	err = cmd.ExecuteContext(context.Background())
	assert.NoError(t, err)
	assert.Contains(t, b.String(), "Tool: my-tool")
	assert.Contains(t, b.String(), "Hash:")
}

func TestToolHashCmd_NotFound(t *testing.T) {
	configContent := `
upstream_services:
  - name: "my-service"
    http_service:
      address: "http://example.com"
      tools:
        - name: "my-tool"
`
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := newRootCmd(fs)
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"tool", "hash", "other-tool", "--config-path", "/config.yaml"})

	err = cmd.ExecuteContext(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestToolHashCmd_InvalidConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/config.yaml", []byte("invalid: yaml: content"), 0644)
	require.NoError(t, err)

	cmd := newRootCmd(fs)
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"tool", "hash", "my-tool", "--config-path", "/config.yaml"})

	err = cmd.ExecuteContext(context.Background())
	assert.Error(t, err)
	// The error comes from config load which calls viper read which might fail or unmarshal fail
	// Actually, config.LoadService calls config.Load which uses viper.
	// If yaml is invalid, it might fail validation or unmarshal.
	// We expect "failed to load configuration" or "configuration load failed"
	assert.Contains(t, err.Error(), "failed")
}

func TestToolHashCmd_MultipleServiceTypes(t *testing.T) {
	configContent := `
upstream_services:
  - name: "http-service"
    http_service:
      address: "http://example.com"
      tools:
        - name: "http-tool"
  - name: "grpc-service"
    grpc_service:
      address: "localhost:50051"
      tools:
        - name: "grpc-tool"
  - name: "cli-service"
    command_line_service:
      command: "echo"
      tools:
        - name: "cli-tool"
`
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	tests := []struct {
		toolName string
	}{
		{"http-tool"},
		{"grpc-tool"},
		{"cli-tool"},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			cmd := newRootCmd(fs)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.SetArgs([]string{"tool", "hash", tt.toolName, "--config-path", "/config.yaml"})

			err = cmd.ExecuteContext(context.Background())
			assert.NoError(t, err)
			assert.Contains(t, b.String(), "Tool: "+tt.toolName)
		})
	}
}
