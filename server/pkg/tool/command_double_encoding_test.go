// Copyright 2025 Author(s) of MCP Any
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

func TestCommandTool_DoubleEncodingVulnerability(t *testing.T) {
	// This test demonstrates a vulnerability where CommandTool (used for non-local or default execution)
	// fails to check for double-encoded path traversal characters, whereas LocalCommandTool correctly checks them.
	// This allows a double-encoded path traversal payload to bypass security checks in CommandTool.

	// Payload: %252e%252e (Double encoded "..")
	// Level 1 (Raw): %252e%252e
	// Level 2 (Decoded once): %2e%2e (Encoded "..")
	// Level 3 (Decoded twice): ..
	payload := "%252e%252e"

	// Define a simple tool that echoes the argument
	toolProto := v1.Tool_builder{
		Name: proto.String("echo-tool"),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"properties": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"arg": structpb.NewStructValue(&structpb.Struct{}),
					},
				}),
			},
		},
	}.Build()

	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
	}).Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{arg}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("arg")}.Build(),
			}.Build(),
		},
	}.Build()

	inputData := map[string]interface{}{"arg": payload}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)
	req := &tool.ExecutionRequest{ToolInputs: inputs, ToolName: "echo-tool"}

	// 1. Verify LocalCommandTool BLOCKS the payload
	// LocalCommandTool uses validateSafePathAndInjection which checks decoded values.
	localTool := tool.NewLocalCommandTool(toolProto, service, callDef, nil, "local-call")
	_, err = localTool.Execute(context.Background(), req)
	assert.Error(t, err, "LocalCommandTool should block double-encoded path traversal")
	if err != nil {
		assert.Contains(t, err.Error(), "path traversal attempt detected", "Error message should mention path traversal")
	}

	// 2. Verify CommandTool BLOCKS the payload (Fix Verified)
	// CommandTool should now use validateSafePathAndInjection which checks decoded values.
	commandTool := tool.NewCommandTool(toolProto, service, callDef, nil, "remote-call")
	_, err = commandTool.Execute(context.Background(), req)

	assert.Error(t, err, "CommandTool should block double-encoded path traversal")
	if err != nil {
		assert.Contains(t, err.Error(), "path traversal attempt detected", "Error message should mention path traversal")
	}
}
