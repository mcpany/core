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

func TestCommandTool_ShellInjection_Prevention(t *testing.T) {
	// Setup a CommandTool configured to use "awk"
	// This test verifies that shell injection protections are active for CommandTool.

	// Define the tool service
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
	}).Build()

	// Define the call definition allowing "args"
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
			}.Build(),
		},
	}.Build()

	// Allow "args" in input schema
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"args": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}

	toolProto := v1.Tool_builder{
		Name:        proto.String("rce-test-awk"),
		InputSchema: inputSchema,
	}.Build()

	cmdTool := tool.NewCommandTool(
		toolProto,
		service,
		callDef,
		nil,
		"call-id",
	)

	// Payload: awk 'BEGIN { system("echo INJECTED") }'
	// We pass the program as the first argument.

	inputData := map[string]interface{}{
		"args": []string{"BEGIN { system(\"echo INJECTED\") }"},
	}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)

	req := &tool.ExecutionRequest{ToolInputs: inputs}

	// Execute
	_, err = cmdTool.Execute(context.Background(), req)

	// Expect an error due to injection detection
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "injection detected")
		t.Log("Security Fix Verified: Injection blocked")
	} else {
		t.Error("Security Fix Failed: Injection was not blocked")
	}
}
