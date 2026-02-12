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

func TestSensitiveFileAccess_Blocked(t *testing.T) {
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

	// Execute the tool requesting access to .env
	req := &ExecutionRequest{
		ToolName: "cat-tool",
		// JSON payload for input
		ToolInputs: []byte(`{"file": ".env"}`),
	}

	_, err = localTool.Execute(context.Background(), req)

	// Expect ERROR now because access to .env is restricted
	assert.Error(t, err, "Should return error when accessing sensitive file")
	if err != nil {
		assert.Contains(t, err.Error(), "restricted", "Error message should mention restriction")
	}
}

func TestSensitiveDirectoryAccess_Blocked(t *testing.T) {
    // Create a dummy .git directory
    tmpDir := t.TempDir()
    gitDir := filepath.Join(tmpDir, ".git")
    err := os.Mkdir(gitDir, 0755)
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

    // Define a tool that uses 'ls' to list a directory
    tool := v1.Tool_builder{
        Name: proto.String("ls-tool"),
    }.Build()

    service := configv1.CommandLineUpstreamService_builder{
        Command: proto.String("ls"),
        Local:   proto.Bool(true),
    }.Build()

    param := configv1.CommandLineParameterMapping_builder{
        Schema: configv1.ParameterSchema_builder{
            Name: proto.String("dir"),
        }.Build(),
    }.Build()

    callDef := configv1.CommandLineCallDefinition_builder{
        Args:       []string{"{{dir}}"},
        Parameters: []*configv1.CommandLineParameterMapping{param},
    }.Build()

    // Create policies (empty)
    policies := []*configv1.CallPolicy{}

    // Create the tool
    localTool := NewLocalCommandTool(tool, service, callDef, policies, "call-id")

    // Execute the tool requesting access to .git
    req := &ExecutionRequest{
        ToolName: "ls-tool",
        // JSON payload for input
        ToolInputs: []byte(`{"dir": ".git"}`),
    }

    _, err = localTool.Execute(context.Background(), req)

    // Expect ERROR now because access to .git is restricted
    assert.Error(t, err, "Should return error when accessing sensitive directory")
    if err != nil {
        assert.Contains(t, err.Error(), "restricted", "Error message should mention restriction")
    }
}
