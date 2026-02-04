// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_SchemaValidation(t *testing.T) {
	t.Parallel()

	// Define a tool that echos the argument
	cmdService := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
	}.Build()

	cmdCallDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{num}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("num"),
					Type: configv1.ParameterType_INTEGER.Enum(),
				}.Build(),
			}.Build(),
		},
	}.Build()

	// Input Schema requiring integer
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"num": map[string]interface{}{"type": "integer"},
		},
	})

	toolProto := v1.Tool_builder{
		Name:        proto.String("echo_num"),
		InputSchema: inputSchema,
	}.Build()

	localTool := tool.NewLocalCommandTool(toolProto, cmdService, cmdCallDef, nil, "call-id")

	// Payload with STRING instead of INTEGER
	inputs := map[string]interface{}{
		"num": "not_an_integer",
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &tool.ExecutionRequest{
		ToolName:   "echo_num",
		ToolInputs: inputBytes,
		DryRun:     true,
	}

	// This must fail validation
	_, err := localTool.Execute(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input validation failed")
	assert.Contains(t, err.Error(), "expected integer, but got string")
}
