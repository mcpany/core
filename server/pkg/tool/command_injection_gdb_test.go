// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGDBInjection(t *testing.T) {
	// This test reproduces a vulnerability where tools not in the blocklist
	// (like gdb) can execute arbitrary shell commands via specific arguments.
	// gdb allows "shell <cmd>" or "!" to execute shell commands.

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
		Description: proto.String("A gdb tool"),
		InputSchema: inputSchema,
	}.Build()

	// Configuration: gdb -batch -ex <user-input>
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("gdb"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("command")}.Build(),
			}.Build(),
		},
		Args: []string{"-batch", "-ex", "{{command}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Attack payload: execute 'echo vulnerable' via gdb shell command
	req := &ExecutionRequest{
		ToolName: "gdb-tool",
		Arguments: map[string]interface{}{
			"command": "shell echo vulnerable",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := localTool.Execute(ctx, req)

	// After fix, this should error out with "shell injection detected"
	assert.Error(t, err)
	assert.Nil(t, result)
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected")
		t.Logf("Execution successfully blocked: %v", err)
	}
}
