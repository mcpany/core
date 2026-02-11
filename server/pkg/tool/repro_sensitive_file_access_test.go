// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestReproSensitiveFileAccess(t *testing.T) {
	// Create a dummy .env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	secretContent := "MCPANY_API_KEY=super-secret-key"
	err := os.WriteFile(envPath, []byte(secretContent), 0600)
	require.NoError(t, err)

	// Save original WD
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// Switch to tmpDir
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Define a tool that uses 'cat' to read a file
	tool := v1.Tool_builder{
		Name: proto.String("cat-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("cat"),
		Local:   proto.Bool(true),
	}.Build()

	// Use CommandLineParameterMapping
	param := configv1.CommandLineParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("file"),
		}.Build(),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args:       []string{"{{file}}"},
		Parameters: []*configv1.CommandLineParameterMapping{param},
	}.Build()

    // Create policies (empty)
    policies := []*configv1.CallPolicy{}

	// Create the tool
	localTool := NewLocalCommandTool(tool, service, callDef, policies, "call-id")

	// Test Case 1: .env
	t.Run("Block .env access", func(t *testing.T) {
		req := &ExecutionRequest{
			ToolName: "cat-tool",
			ToolInputs: []byte(`{"file": ".env"}`),
		}

		result, err := localTool.Execute(context.Background(), req)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "access to sensitive file or directory \".env\" is blocked")
		}
		assert.Nil(t, result)
	})

	// Test Case 2: .git/config
	t.Run("Block .git access", func(t *testing.T) {
		req := &ExecutionRequest{
			ToolName: "cat-tool",
			ToolInputs: []byte(`{"file": ".git/config"}`),
		}

		result, err := localTool.Execute(context.Background(), req)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "access to sensitive file or directory \".git\" is blocked")
		}
		assert.Nil(t, result)
	})

	// Test Case 3: .ssh/id_rsa
	t.Run("Block .ssh access", func(t *testing.T) {
		req := &ExecutionRequest{
			ToolName: "cat-tool",
			ToolInputs: []byte(`{"file": ".ssh/id_rsa"}`),
		}

		result, err := localTool.Execute(context.Background(), req)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "access to sensitive file or directory \".ssh\" is blocked")
		}
		assert.Nil(t, result)
	})

    // Test Case 4: config.yaml
	t.Run("Block config.yaml access", func(t *testing.T) {
		req := &ExecutionRequest{
			ToolName: "cat-tool",
			ToolInputs: []byte(`{"file": "config.yaml"}`),
		}

		result, err := localTool.Execute(context.Background(), req)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "access to sensitive file or directory \"config.yaml\" is blocked")
		}
		assert.Nil(t, result)
	})

    // Test Case 5: Safe file
	t.Run("Allow safe file access", func(t *testing.T) {
        safeFile := "safe.txt"
        os.WriteFile(safeFile, []byte("safe content"), 0600)
		req := &ExecutionRequest{
			ToolName: "cat-tool",
			ToolInputs: []byte(`{"file": "safe.txt"}`),
		}

		result, err := localTool.Execute(context.Background(), req)
		assert.NoError(t, err)
        if err == nil {
            // Verify content if possible, but execute might return something different depending on implementation (stdout map)
            resultMap, ok := result.(map[string]interface{})
            require.True(t, ok)
            stdout, ok := resultMap["stdout"].(string)
            require.True(t, ok)
            assert.Contains(t, stdout, "safe content")
        }
	})
}
