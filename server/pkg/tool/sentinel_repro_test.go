// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestInterpreterInjection_Bypass(t *testing.T) {
	// Reproduces a bypass where analyzeQuoteContext thinks the input is in single quotes (Level 2)
	// because of the outer shell command, but the inner command (python) interprets double quotes.
	// This allows injecting double quotes and breaking out of the python string, then using
	// module functions like subprocess.run to execute code.

	cmd := "bash"
	// The template uses single quotes for bash argument, but inside it uses double quotes for python string.
	// analyzeQuoteContext sees the outer single quotes and sets Level 2.
	// Level 2 blocks single quotes but allows double quotes.
	// Python sees double quotes and allows breaking out with ".
	template := "python3 -c 'print(\"{{input}}\")'"

	toolDef := v1.Tool_builder{Name: proto.String("test-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", template},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()
	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Payload attempts to:
	// 1. Break out of python string with "
	// 2. Import subprocess (if not already available, but here we assume we can import or use __import__)
	// 3. Execute code using subprocess.run
	// Note: We use subprocess.run because checkInterpreterFunctionCalls checks for `subprocess(`
	// but currently misses `subprocess.`.

	// Sentinel Update: We must avoid single quotes because analyzeQuoteContext (Level 2) blocks them.
	// We use double quotes inside the python code.
	payload := `"); import subprocess; subprocess.run(["echo", "pwned"]); print("`
	// Escape double quotes for JSON string
	jsonPayload := strings.ReplaceAll(payload, `"`, `\"`)

	req := &ExecutionRequest{
		ToolName: "test",
		ToolInputs: []byte(`{"input": "` + jsonPayload + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)

	// After the fix, we expect it to FAIL with "interpreter injection detected".
	if err != nil {
		t.Logf("Detected error (expected): %v", err)
		assert.Contains(t, err.Error(), "interpreter injection detected")
	} else {
		t.Logf("Vulnerability Reproduced: Injection was NOT detected!")
		t.Fail()
	}
}

func TestInterpreterInjection_AliasBypass(t *testing.T) {
	// Reproduces a bypass of the previous fix (blocking subprocess.) by using import aliasing.
	// This shows that blacklisting specific module names is insufficient if we can still break out of the string.

	cmd := "bash"
	template := "python3 -c 'print(\"{{input}}\")'"

	toolDef := v1.Tool_builder{Name: proto.String("test-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", template},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()
	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Payload attempts to bypass the "subprocess." block by aliasing it to 's'.
	// Then calls s.run().
	// This requires breaking out of the double-quoted python string first (using ").
	payload := `"); import subprocess as s; s.run(["echo", "pwned"]); print("`
	// Escape double quotes for JSON string
	jsonPayload := strings.ReplaceAll(payload, `"`, `\"`)

	req := &ExecutionRequest{
		ToolName: "test",
		ToolInputs: []byte(`{"input": "` + jsonPayload + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)

	// After the root cause fix (blocking "), this should FAIL with "value contains double quote".
	if err != nil {
		t.Logf("Detected error (expected): %v", err)
		assert.Contains(t, err.Error(), "value contains double quote")
	} else {
		t.Logf("Vulnerability Reproduced: Injection was NOT detected!")
		t.Fail()
	}
}
