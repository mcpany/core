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

func TestCommandTool_AwkRCE_Vulnerability(t *testing.T) {
	// Setup a CommandTool configured to use "awk"
	// We force CommandTool usage (no local: true, no container env).

	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
	}).Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
			}.Build(),
		},
	}.Build()

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
		Name:        proto.String("awk-rce-test"),
		InputSchema: inputSchema,
	}.Build()

	cmdTool := tool.NewCommandTool(
		toolProto,
		service,
		callDef,
		nil,
		"call-id",
	)

	// Payload: Use awk's system() function to execute arbitrary command.
	// Input: ["BEGIN { system(\"echo INJECTED\") }"]

	inputData := map[string]interface{}{
		"args": []string{"BEGIN { system(\"echo INJECTED\") }"},
	}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)

	req := &tool.ExecutionRequest{ToolInputs: inputs}

	// Execute
	result, err := cmdTool.Execute(context.Background(), req)

	// Expectation: The execution should fail because of the security check.
	require.Error(t, err)
	assert.Nil(t, result)
	// It might fail with "shell injection detected" (due to unquoted chars) or "interpreter injection detected".
	assert.Contains(t, err.Error(), "injection detected")
}
