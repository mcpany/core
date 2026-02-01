// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_InterpreterInjection_Python(t *testing.T) {
	// Define a tool that runs python3 -c 'print("{{msg}}")'
	tool := v1.Tool_builder{
		Name: proto.String("python-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('{{msg}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: hello'); print('INJECTED'); #
	// This closes the single quote, runs some code, and comments out the rest.
	// The validation logic checks for "system(", "exec(", "popen(", "eval(".
	// It does NOT check for "print" or "open".
	payload := "hello'); print('INJECTED'); #"

	inputs := map[string]interface{}{
		"msg": payload,
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "python-tool",
		ToolInputs: json.RawMessage(inputBytes),
	}

	_, err := localTool.Execute(context.Background(), req)

	// Ideally, this should fail security validation.
	if err != nil {
		t.Logf("Blocked as expected. Error: %v", err)
		if strings.Contains(err.Error(), "injection detected") {
			t.Log("Security check caught it!")
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	t.Fatal("Python injection was NOT blocked!")
}
