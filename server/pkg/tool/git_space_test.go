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

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
			}.Build(),
		},
	}.Build()

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
		v1.Tool_builder{InputSchema: inputSchema}.Build(),
		service,
		callDef,
		nil,
		"call-id",
	)

	// Try to execute: git status "my folder"
	// We pass "status", "my folder"
	// This should now SUCCEED (or at least pass the security check)

	inputData := map[string]interface{}{"args": []string{"status", "my folder"}}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	_, err = gitTool.Execute(context.Background(), req)

	// We expect NO error related to shell injection.
	// It might fail because "my folder" doesn't exist or not a git repo, but that's fine.
	// If it fails with "shell injection detected", then the fix is not working.
	if err != nil {
		assert.NotContains(t, err.Error(), "shell injection detected")
	}
}

func TestShellInjectionStillBlocked(t *testing.T) {
	// Ensure that actual shells still block dangerous characters

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
			}.Build(),
		},
	}.Build()

	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
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

	shTool := tool.NewLocalCommandTool(
		v1.Tool_builder{InputSchema: inputSchema}.Build(),
		service,
		callDef,
		nil,
		"call-id",
	)

	// Try to execute: sh "ls; rm -rf /"
	// We pass "ls; rm -rf /" as argument.
	// This is not a flag, so argument injection check passes.
	// But it is a shell command ("sh"), so shell injection check runs.
	// "ls; rm -rf /" contains ";" which is blocked in unquoted context.

	inputData := map[string]interface{}{"args": []string{"ls; rm -rf /"}}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	_, err = shTool.Execute(context.Background(), req)

	// Expect error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shell injection detected")
}
