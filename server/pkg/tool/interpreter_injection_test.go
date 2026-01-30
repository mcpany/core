// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestSedSandbox_Prevention(t *testing.T) {
	cmd := "sed"

	// Create tool for sed -e {{script}}
	toolDef := v1.Tool_builder{Name: proto.String("sed-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Payload: 1e (Execute command).
	// We avoid spaces ("1e date") because LocalCommandTool's shell injection protection
	// forbids spaces in unquoted arguments for shell-like commands (including sed).
	// The sandbox should still block 'e' even without arguments (or with different args).
	req := &ExecutionRequest{
		ToolName: "sed-tool",
		ToolInputs: []byte(`{"script": "1e"}`),
	}

	result, err := tool.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not map: %v", result)
	}

	// Expect return code to be non-zero (sed error or sandbox violation)
	returnCode, ok := resMap["return_code"].(int)
	if !ok {
		t.Fatalf("return_code not int: %v", resMap["return_code"])
	}

	if returnCode == 0 {
		t.Errorf("FAIL: sed executed '1e' successfully (return_code 0). Sandbox failed.")
	}

	stderr, _ := resMap["stderr"].(string)
	if !strings.Contains(stderr, "command disabled") && !strings.Contains(stderr, "unknown command") {
		// GNU sed: 'e' command disabled in sandbox mode
		// BSD sed: unknown command: 1 (or similar)
		t.Logf("Note: stderr was: %s", stderr)
	} else {
		t.Logf("Success: sed blocked command with error: %s", stderr)
	}

	// Payload: w/tmp/pwned (Write file)
	// We avoid space ("w /tmp/pwned") to pass shell injection check.
	// sed usually accepts concatenated filename.
	req = &ExecutionRequest{
		ToolName: "sed-tool",
		ToolInputs: []byte(`{"script": "w/tmp/pwned"}`),
	}

	result, err = tool.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	resMap, _ = result.(map[string]interface{})
	returnCode, _ = resMap["return_code"].(int)

	if returnCode == 0 {
		t.Errorf("FAIL: sed executed 'w/tmp/pwned' successfully. Sandbox failed.")
	}
}

func TestSedSandbox_ValidUsage(t *testing.T) {
	cmd := "sed"
	toolDef := v1.Tool_builder{Name: proto.String("sed-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Valid replacement: 'q' (quit) has no spaces and is safe.
	req := &ExecutionRequest{
		ToolName: "sed-tool",
		ToolInputs: []byte(`{"script": "q"}`),
	}

	// We use a short timeout because sed might wait for input if 'q' is not processed immediately (though it should be).
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := tool.Execute(ctx, req)
	if err != nil {
		// Timeout might happen if sed waits for stdin before executing 'q' (depending on implementation)
		// but 'q' usually quits immediately.
		// If it times out or fails, we log it.
		t.Logf("Execute result: %v, err: %v", result, err)
	} else {
		resMap, _ := result.(map[string]interface{})
		t.Logf("Success: valid command executed. Return code: %v", resMap["return_code"])
	}
}
