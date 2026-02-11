// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func strPtrInj(s string) *string { return &s }

func TestPythonDoubleQuoteInjection(t *testing.T) {
	// This test reproduces an RCE vulnerability where Python code can be injected
	// into a double-quoted argument string.

	service := configv1.CommandLineUpstreamService_builder{
		Command: strPtrInj("python3"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print(\"Hello, {{name}}!\")"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: strPtrInj("name"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := pb.Tool_builder{
		Name: strPtrInj("python_hello"),
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload: Close the print function, inject import sys; sys.exit(42), and comment out the rest.
	// If this executes, the exit code will be 42.
	// Payload: user'); import sys; sys.exit(42); #
	payload := "user'); import sys; sys.exit(42); #"

	req := &ExecutionRequest{
		ToolName: "python_hello",
		ToolInputs: []byte(`{"name": "` + payload + `"}`),
	}

	result, err := tool.Execute(context.Background(), req)

	if err != nil {
		t.Logf("Execution blocked or failed: %v", err)
		// If exit code is 42, it means RCE successful
		// The error message for non-zero exit code usually contains the code
		if assert.NotContains(t, err.Error(), "exit status 42") {
             t.Log("Safe: Exit code was not 42")
        } else {
             t.Fatal("VULNERABILITY CONFIRMED: Python code injection successful (exit code 42)")
        }
	} else {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		returnCode, ok := resMap["return_code"].(int)
		// return_code might be int or float64 depending on unmarshaling
        if !ok {
             // check float64
             if rcFloat, ok := resMap["return_code"].(float64); ok {
                 returnCode = int(rcFloat)
             } else {
                 // Might be 0 if not present?
                 returnCode = 0
             }
        }

		t.Logf("Return Code: %d", returnCode)

		if returnCode == 42 {
			t.Fatal("VULNERABILITY CONFIRMED: Python code injection successful (exit code 42)")
		} else {
            t.Log("Safe: Exit code was not 42")
        }

        stdout, _ := resMap["stdout"].(string)
        t.Logf("Stdout: %s", stdout)
	}
}

func TestLocalCommandTool_PythonEvalInjection_FunctionAlias(t *testing.T) {
	// This test demonstrates that Python interpreter injection can bypass
	// function call checks by aliasing dangerous functions, even inside
	// an exec/eval context which the security layer attempts to protect.

	t.Parallel()

	toolProto := (&pb.Tool_builder{Name: proto.String("python-eval-tool")}).Build()
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		Local:   proto.Bool(true),
	}).Build()

	// Pattern: python -c 'exec("{{input}}")'
	// The config uses exec(), effectively allowing code execution,
	// BUT the security layer attempts to filter dangerous calls like system().
	callDef := (&configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "exec(\"{{input}}\")"},
		Parameters: []*configv1.CommandLineParameterMapping{
			(&configv1.CommandLineParameterMapping_builder{
				Schema: (&configv1.ParameterSchema_builder{Name: proto.String("input")}).Build(),
			}).Build(),
		},
	}).Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload: import os; s=os.system; s('echo injected')
	// Uses single quotes to bypass Level 1 (Double Quoted) check which blocks "
	// Uses aliasing to bypass checkInterpreterFunctionCalls which blocks system(
	payload := "import os; s=os.system; s('echo injected')"

	args := map[string]interface{}{
		"input": payload,
	}
	inputs, _ := json.Marshal(args)
	req := &ExecutionRequest{
		ToolName:   "python-eval-tool",
		ToolInputs: inputs,
	}

	result, err := localTool.Execute(context.Background(), req)

	// We expect the security layer to BLOCK this injection.
	// So err should be non-nil and contain detection message.
	assert.Error(t, err, "Expected security error, got nil")
	if err != nil {
		assert.Contains(t, err.Error(), "interpreter injection detected", "Error should indicate interpreter injection")
	}

	if result != nil {
		resMap, ok := result.(map[string]interface{})
		if ok {
			var output string
			if s, ok := resMap["stdout"].(string); ok {
				output += s
			}
			if s, ok := resMap["combined_output"].(string); ok {
				output += s
			}
			if strings.Contains(output, "injected") {
				t.Errorf("VULNERABILITY CONFIRMED: Output contains 'injected'")
			}
		}
	}
}
