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

func TestLocalCommandTool_ArgumentInjection_Prevention(t *testing.T) {
	// This test verifies that argument injection via placeholders is prevented.

	tool := &v1.Tool{
		Name:        proto.String("test-tool-cat"),
	}
	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("cat"),
		Local:   proto.Bool(true),
	}
	callDef := &configv1.CommandLineCallDefinition{
		Parameters: []*configv1.CommandLineParameterMapping{
			{Schema: &configv1.ParameterSchema{Name: proto.String("file")}},
		},
		Args: []string{"{{file}}"},
	}

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Case 1: Safe input
	reqSafe := &ExecutionRequest{
		ToolName: "test-tool-cat",
		Arguments: map[string]interface{}{
			"file": "/etc/hosts",
		},
	}
	reqSafe.ToolInputs, _ = json.Marshal(reqSafe.Arguments)

	_, err := localTool.Execute(context.Background(), reqSafe)
	assert.NoError(t, err)

	// Case 2: Argument Injection
	reqAttack := &ExecutionRequest{
		ToolName: "test-tool-cat",
		Arguments: map[string]interface{}{
			"file": "-n", // Attempt to inject a flag
		},
	}
	reqAttack.ToolInputs, _ = json.Marshal(reqAttack.Arguments)

	_, err = localTool.Execute(context.Background(), reqAttack)

	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "argument injection")
	}

	// Case 3: Negative number (should be allowed)
	reqNegative := &ExecutionRequest{
		ToolName: "test-tool-cat",
		Arguments: map[string]interface{}{
			"file": "-5",
		},
	}
	reqNegative.ToolInputs, _ = json.Marshal(reqNegative.Arguments)

	// We expect this to fail with "argument injection" if we block all negatives,
	// or pass if we allow numbers.
	// Our implementation allows numbers.
	// However, "cat -5" will fail with "No such file", but NOT "argument injection".
	_, err = localTool.Execute(context.Background(), reqNegative)

	// Error could be from cat failing, or injection check.
	// We want to ensure it is NOT injection check.
	if err != nil {
		assert.NotContains(t, err.Error(), "argument injection")
	}
}

func TestLocalCommandTool_ArgsParameter_Injection_Prevention(t *testing.T) {
	// This test verifies that argument injection via the "args" parameter is prevented.
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
		Name:        proto.String("test-tool-args-vuln"),
		Description: proto.String("A test tool vulnerability"),
		InputSchema: inputSchema,
	}
	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}
	callDef := &configv1.CommandLineCallDefinition{
		Parameters: []*configv1.CommandLineParameterMapping{
			{Schema: &configv1.ParameterSchema{Name: proto.String("args")}},
		},
	}

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Case: Pass a flag as an argument in the 'args' list
	req := &ExecutionRequest{
		ToolName: "test-tool-args-vuln",
		Arguments: map[string]interface{}{
			"args": []interface{}{"-n", "hello"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	assert.Error(t, err, "Expected error due to argument injection")
	if err != nil {
		assert.Contains(t, err.Error(), "argument injection detected")
	}
}
