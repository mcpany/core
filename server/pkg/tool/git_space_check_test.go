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
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGitSpaceInjectionRepro(t *testing.T) {
	// This test reproduces the issue where arguments with spaces are blocked for "git"
	// even though git handles them safely.

	// Setup a tool definition for "git"
	callDef := (&configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			(&configv1.CommandLineParameterMapping_builder{
				Schema: (&configv1.ParameterSchema_builder{Name: proto.String("args")}).Build(),
			}).Build(),
		},
	}).Build()

	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
	}).Build()

	properties := map[string]*structpb.Value{
		"args": structpb.NewStructValue(&structpb.Struct{}),
	}
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: properties,
			}),
		},
	}

	gitTool := tool.NewLocalCommandTool(
		(&v1.Tool_builder{InputSchema: inputSchema}).Build(),
		service,
		callDef,
		nil,
		"call-id",
	)

	// Try to execute: git status "my folder"
	inputData := map[string]interface{}{"args": []string{"status", "my folder"}}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	// This should succeed (return a result, not an error)
	res, err := gitTool.Execute(context.Background(), req)

	// We expect validation to pass, so err should be nil.
	// The command execution itself might fail (exit code != 0), but that's returned in res.
	assert.NoError(t, err, "Validation failed unexpectedly (shell injection error?)")

	if resMap, ok := res.(map[string]interface{}); ok {
		t.Logf("Execution result: %+v", resMap)
	}
}
