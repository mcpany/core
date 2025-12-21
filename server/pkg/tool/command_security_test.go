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

func TestLocalCommandTool_Execute_ArgsInjection(t *testing.T) {
	// Define a tool that allows 'args' parameter
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"args": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := &v1.Tool{
		Name:        proto.String("test-tool-args"),
		Description: proto.String("A test tool enabling args"),
		InputSchema: inputSchema,
	}
	service := &configv1.CommandLineUpstreamService{}
	service.Command = proto.String("echo")
	service.Local = proto.Bool(true)
	// No defined mappings, just relying on args
	callDef := &configv1.CommandLineCallDefinition{}

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Attempt to pass a flag via args
	req := &ExecutionRequest{
		ToolName: "test-tool-args",
		Arguments: map[string]interface{}{
			"args": []interface{}{"--injected-flag", "value"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	_, err := localTool.Execute(context.Background(), req)

	// Expectation: This SHOULD fail with an argument injection error
	assert.Error(t, err, "Should return error for argument starting with -")
    if err != nil {
	    assert.Contains(t, err.Error(), "argument injection detected")
    }
}
