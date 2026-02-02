// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestGitCommitWithSpace(t *testing.T) {
	// Define a git tool
	toolProto := v1.Tool_builder{
		Name: proto.String("git_commit"),
	}.Build()

	serviceConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		// Env is optional, so we can omit it
	}.Build()

	callDefinition := configv1.CommandLineCallDefinition_builder{
		Args: []string{"commit", "-m", "{{message}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("message"),
					Type: configv1.ParameterType_STRING.Enum(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	// Use NewLocalCommandTool directly
	localTool := tool.NewLocalCommandTool(toolProto, serviceConfig, callDefinition, nil, "call1")

	// Input with space
	inputData := map[string]interface{}{"message": "initial commit"}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)

	req := &tool.ExecutionRequest{
		ToolName:   "git_commit",
		ToolInputs: inputs,
	}

	// This should succeed now as "git" is removed from isShellCommand list
	// and even if it was there, space is removed from dangerousChars.
	result, err := localTool.Execute(context.Background(), req)

	// We expect success now
	require.NoError(t, err)

	// Since we are running a real git command, it might fail if git is not installed or not in a repo.
	// But localTool.Execute returns the result map even on failure (status=error).
	// We just want to ensure it didn't fail with "shell injection detected".

	// Check that we got a result
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	// The command execution itself might fail (e.g. not a git repo), but that's fine.
	// We verify that the command "git" was attempted.
	assert.Equal(t, "git", resultMap["command"])
}
