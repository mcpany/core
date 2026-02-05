// Copyright 2025 Author(s) of MCP Any
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

func TestLocalCommandTool_ArgsInjection_Vim(t *testing.T) {
	// This test attempts to inject a malicious editor command via the 'args' parameter
	// into a 'vim' command. 'vim' supports '+cmd' arguments to execute Ex commands.
	// After the fix, it should be blocked.

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
		Name:        proto.String("vim-tool"),
		Description: proto.String("A dangerous vim tool"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("vim"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "vim-tool",
		Arguments: map[string]interface{}{
			// The injection: passing a + command.
			"args": []interface{}{"+so", "missing_file.vim"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute the tool
	_, err := localTool.Execute(context.Background(), req)

	// With the fix, we expect a security error.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "argument injection detected: value starts with '+'")
}

func TestLocalCommandTool_ArgsInjection_Awk(t *testing.T) {
	// Verify that awk script injection is also blocked (by existing checks, or new ones for CommandTool)

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
		Name:        proto.String("awk-tool"),
		Description: proto.String("A dangerous awk tool"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"args": []interface{}{"BEGIN { print \"pwned\" }"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute the tool
	_, err := localTool.Execute(context.Background(), req)

	assert.Error(t, err)
	// Expect shell injection detected because of { or space
	// "shell injection detected: value contains dangerous character '{'"
	// or similar.
	assert.Contains(t, err.Error(), "shell injection detected")
}
