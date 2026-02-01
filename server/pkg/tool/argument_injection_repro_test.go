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

func TestLocalCommandTool_ArgumentInjection_ResponseFile(t *testing.T) {
	t.Parallel()
	// Define a tool that uses 'gcc', which supports @file syntax.
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"flags": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("gcc-tool"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("gcc"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `gcc {{flags}}`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{flags}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("flags")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// We attempt to pass a response file argument '@malicious_args'
	// This should be blocked to prevent argument injection via file.
	req := &ExecutionRequest{
		ToolName: "gcc-tool",
		Arguments: map[string]interface{}{
			"flags": "@malicious_args",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// We expect this to fail due to the new security check
	result, err := localTool.Execute(context.Background(), req)

	// Verify that the vulnerability is patched
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "argument injection detected")
	assert.Contains(t, err.Error(), "response file")
	assert.Nil(t, result)
}
