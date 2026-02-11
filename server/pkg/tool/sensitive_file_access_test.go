// Copyright 2026 Author(s) of MCP Any
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

	result, err := localTool.Execute(context.Background(), req)

	// Sentinel Security Update: We expect this to fail now due to IsSensitivePath check.
	if err != nil {
		t.Logf("Secure: Execution failed as expected with error: %v", err)
		assert.Contains(t, err.Error(), "access to environment file \".env\" is blocked")
		return
	}

	// If we reach here, it means the execution succeeded, which is a failure of security.
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	stdout, ok := resultMap["stdout"].(string)
	require.True(t, ok)

	if stdout == secretContent {
		t.Errorf("Vulnerability Confirmed: Successfully read .env file content: %s", stdout)
	} else {
		t.Logf("Output: %s", stdout)
	}

	assert.Fail(t, "Should not be able to execute tool with .env file access")
}
