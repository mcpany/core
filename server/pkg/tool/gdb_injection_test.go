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

func TestLocalCommandTool_GDBInjection_Blocked(t *testing.T) {
	t.Parallel()
	// Define a tool that uses 'gdb', which IS in isShellCommand list.
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"command": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("gdb-tool"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("gdb"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `gdb -batch -ex {{command}}`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-batch", "-ex", "{{command}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("command")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// We attempt to pass a command with dangerous characters (semicolon)
	// Since gdb is now detected as a shell, this should raise "shell injection detected"
	req := &ExecutionRequest{
		ToolName: "gdb-tool",
		Arguments: map[string]interface{}{
			"command": "shell echo pwned; ls",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)

	// Expect error due to shell injection detection
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "shell injection detected")
	}
	assert.Nil(t, result)
}
