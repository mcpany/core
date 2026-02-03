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
)

func TestLocalCommandTool_PythonSubprocessInjection(t *testing.T) {
	// Setup a tool that runs python with a single-quoted argument
	tool := v1.Tool_builder{
		Name: proto.String("python-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		Local:   proto.Bool(true),
	}.Build()

	// python3 -c "exec('{{msg}}')"
	// We use single quotes inside double quotes (conceptually for Python, but strict string for Go/Exec)
	// Template: exec('{{msg}}')
	// quoteLevel = 2 (Single Quoted)
	// This allows spaces in the input!
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "exec('{{msg}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Context: exec('{{msg}}')
	// quoteLevel = 2 (Single Quoted)
	// Blocks: ' (single quote), ` (backtick)
	// Blocks: system(, exec(, popen(, eval(

	// Payload: import subprocess; subprocess.call(["echo", "hacked"])
	// We use double quotes for strings inside the payload (allowed in Single Quoted context).
	// We MUST escape the double quotes because they are inside a JSON string.
	payload := `import subprocess; subprocess.call([\"echo\", \"hacked\"])`

	req := &ExecutionRequest{
		ToolName: "python-tool",
		ToolInputs: json.RawMessage(`{"msg": "` + payload + `"}`),
	}

	// We expect the validation to catch the injection.
	result, err := localTool.Execute(context.Background(), req)

	// Expect failure due to shell injection detection
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected")
	}
	assert.Nil(t, result)
}

func TestLocalCommandTool_FalsePositive_Subprocess(t *testing.T) {
	// Setup a tool that runs python with a single-quoted argument
	tool := v1.Tool_builder{
		Name: proto.String("echo-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}.Build()

	// echo '{{msg}}'
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{msg}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: "I found a subprocess issue"
	// This should PASS because "subprocess" (without dot or import) is safe.
	payload := `I found a subprocess issue`

	req := &ExecutionRequest{
		ToolName: "echo-tool",
		ToolInputs: json.RawMessage(`{"msg": "` + payload + `"}`),
	}

	result, err := localTool.Execute(context.Background(), req)

	// Expect success (validation pass).
	// Execution might succeed or fail depending on if echo is found/env, but err should NOT be "shell injection detected".
	if err != nil {
		assert.NotContains(t, err.Error(), "shell injection detected")
	} else {
		// If execution succeeds, result should be non-nil
		assert.NotNil(t, result)
	}
}
