// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Python_Space_Issue(t *testing.T) {
	// This test demonstrates a usability issue where space is blocked in arguments
	// for commands considered as shells/interpreters, even though exec.Command handles them safely.

	tool := configv1.ToolDefinition_builder{
		Name: proto.String("python-print"),
	}.Build()

	toolProto := v1.Tool_builder{
		Name: proto.String("python-print"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		Local:   proto.Bool(true),
		Tools:   []*configv1.ToolDefinition{tool},
	}.Build()

	// python3 -c "print('{{msg}}')"
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('{{msg}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("msg"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload with space
	payload := "Hello World"

	req := &ExecutionRequest{
		ToolName: "python-print",
		Arguments: map[string]interface{}{
			"msg": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// This currently fails because ' is detected in checkSingleQuotedInjection (Wait, does it?)
	// Template is "print('{{msg}}')". {{msg}} is inside ''.
	// analyzeQuoteContext should return 2 (Single Quoted).
	// checkSingleQuotedInjection should be called.

	if err != nil {
		t.Errorf("Execution failed with valid input: %v", err)
	}
}

func TestLocalCommandTool_Python_Arg_Space_Issue(t *testing.T) {
	// This test uses python as a script runner

	tool := configv1.ToolDefinition_builder{
		Name: proto.String("python-script"),
	}.Build()

	toolProto := v1.Tool_builder{
		Name: proto.String("python-script"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		Local:   proto.Bool(true),
		Tools:   []*configv1.ToolDefinition{tool},
	}.Build()

	// python3 script.py {{arg}}
	// We mock script execution by using -c to print argv
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "import sys; print(sys.argv[1])", "{{arg}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("arg"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload with space
	payload := "Hello World"

	req := &ExecutionRequest{
		ToolName: "python-script",
		Arguments: map[string]interface{}{
			"arg": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// Template is "{{arg}}". Unquoted.
	// checkUnquotedInjection is called.
	// It blocks space.

	if err != nil {
		t.Errorf("Execution failed with valid input containing space: %v", err)
	}
}
