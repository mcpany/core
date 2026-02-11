// Copyright 2026 Author(s) of MCP Any
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

func TestLocalCommandTool_PythonInjection_FunctionAlias(t *testing.T) {
	// This test demonstrates that Python interpreter injection can bypass
	// function call checks by aliasing dangerous functions.

	// Typical pattern: python -c 'print("{{input}}")'
	// The config wraps the input in double quotes.
	callDef := (&configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print(\"{{input}}\")"},
		Parameters: []*configv1.CommandLineParameterMapping{
			(&configv1.CommandLineParameterMapping_builder{
				Schema: (&configv1.ParameterSchema_builder{Name: proto.String("input")}).Build(),
			}).Build(),
		},
	}).Build()

	toolProto := (&v1.Tool_builder{Name: proto.String("python-tool")}).Build()
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		Local:   proto.Bool(true),
	}).Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload: Break out of print("..."), import os, alias system, call alias
	// Input: hello"); import os; s=os.system; s("echo injected"); print("
	// Result: print("hello"); import os; s=os.system; s("echo injected"); print("")
	// Note: We need to escape quotes for the Go string literal
	payload := "hello\"); import os; s=os.system; s(\"echo injected\"); print(\""

	args := map[string]interface{}{
		"input": payload,
	}
	inputs, _ := json.Marshal(args)
	req := &ExecutionRequest{
		ToolName:   "python-tool",
		ToolInputs: inputs,
	}

	result, err := localTool.Execute(context.Background(), req)

	// If vulnerable, this will execute successfully (or fail with execution error but NOT security error)
	// If secure, it should return an error containing "interpreter injection detected"
	if err != nil {
		if strings.Contains(err.Error(), "interpreter injection detected") {
			t.Log("Secure: Injection detected")
		} else {
			t.Errorf("Vulnerable: Execution failed with non-security error: %v", err)
		}
	} else {
		// Execution succeeded -> Vulnerable
		t.Logf("Vulnerable: Execution succeeded with payload: %s", payload)
		if resMap, ok := result.(map[string]interface{}); ok {
			// Check stdout/combined_output
			var output string
			if s, ok := resMap["stdout"].(string); ok {
				output += s
			}
			if s, ok := resMap["combined_output"].(string); ok {
				output += s
			}
			t.Logf("Output: %s", output)
			if strings.Contains(output, "injected") {
				t.Errorf("CONFIRMED RCE: Output contains 'injected'")
			}
		} else {
			t.Errorf("Result is not a map: %T", result)
		}
	}
}
