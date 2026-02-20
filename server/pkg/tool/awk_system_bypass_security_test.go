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

func TestLocalCommandTool_AwkInjection_CommentBypass_System(t *testing.T) {
	// Define a tool that uses 'awk'.
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"val": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("awk-tool-quoted"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `awk` with a quoted template argument
	// awk -v x="{{val}}" ...
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-v", "x=\"{{val}}\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("val")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: 1; //; system(x)
    // The '//' acts as a comment for stripInterpreterComments, causing it to remove 'system(x)'.
    // But awk executes it because '//' is a valid regex.
    // This test ensures that the comment stripping logic correctly handles awk syntax
    // and blocks this bypass.
	req := &ExecutionRequest{
		ToolName: "awk-tool-quoted",
		Arguments: map[string]interface{}{
			"val": "1; //; system(x)",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// We assert that the execution should be BLOCKED with a security error.
    // Before fix, this will fail (err is nil).
    // After fix, this will pass (err contains "interpreter injection").
	if err == nil {
		t.Logf("Execution successful (VULNERABLE): %v", err)
        t.Fail()
	} else {
		t.Logf("Execution failed (SECURE): %v", err)
		assert.Contains(t, err.Error(), "interpreter injection detected", "Should detect interpreter injection")
	}
}

func TestLocalCommandTool_LuaInjection_CommentBypass(t *testing.T) {
	// Define a tool that uses 'lua'.
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"val": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("lua-tool"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("lua"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "x=\"{{val}}\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("val")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: 1 // 2; os.execute("id")
    // If stripInterpreterComments strips // 2..., it strips os.execute.
    // Lua 5.3+ supports // for division.
	req := &ExecutionRequest{
		ToolName: "lua-tool",
		Arguments: map[string]interface{}{
			"val": "1 // 2; os.execute(\"id\")",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	if err == nil {
		t.Logf("Execution successful (VULNERABLE): %v", err)
        t.Fail()
	} else {
		t.Logf("Execution failed (SECURE): %v", err)
		assert.Contains(t, err.Error(), "interpreter injection detected", "Should detect interpreter injection")
	}
}
