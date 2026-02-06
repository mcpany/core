// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_Flags_Allowed_For_Standard_Tools(t *testing.T) {
	t.Parallel()
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
		Name:        proto.String("echo-tool"),
		InputSchema: inputSchema,
	}.Build()

	// 'echo' is NOT in isShellCommand list.
	// So it should accept flags like '-n'.
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "echo-tool",
		Arguments: map[string]interface{}{
			"args": []interface{}{"-n", "hello"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	// echo -n hello -> "hello" (no newline)
	assert.Equal(t, "hello", resultMap["stdout"])
}

func TestLocalCommandTool_Flags_Blocked_For_Shells(t *testing.T) {
	t.Parallel()
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
		Name:        proto.String("sh-tool"),
		InputSchema: inputSchema,
	}.Build()

	// 'sh' IS in isShellCommand list.
	// So it should BLOCK flags like '-c'.
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "sh-tool",
		Arguments: map[string]interface{}{
			// Attempt to inject command via -c
			"args": []interface{}{"-c", "echo pwned"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "argument injection detected: value starts with '-'")
}
