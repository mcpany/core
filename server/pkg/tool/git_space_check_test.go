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
	// This test originally reproduced the issue where arguments with spaces were blocked for "git".
	// Sentinel Security Update: We NOW enforce blocking spaces for git because allowing them enables RCE via configuration injection.
	// See TestLocalCommandTool_Git_C_Injection_Repro for the attack vector.

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

	// This should FAIL now.
	_, err = gitTool.Execute(context.Background(), req)

	// We expect validation to fail.
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected")
		assert.Contains(t, err.Error(), "dangerous character ' '")
	}
}
