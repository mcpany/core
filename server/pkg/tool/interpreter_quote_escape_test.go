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

// TestPythonInjection_MultiParam_Escape checks if using a backslash in a single-quoted parameter
// is properly escaped to prevent breaking out of the string context.
// This simulates: print('{{A}}', '{{B}}')
// A = \
// B = , __import__("os").system("echo POwned")) #
// Result with escaping: print(' \\ ', ' , __import__("os").system("echo POwned")) # ')
// This should print the literal strings and NOT execute the code.
func TestPythonInjection_MultiParam_Escape(t *testing.T) {
	cmd := "python3"

	// Check if python3 is available
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	// simple check - we just assume it exists for now or the test will fail with "executable file not found"
	_ = ctx

	toolDef := v1.Tool_builder{Name: proto.String("python-pwn")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &cmd,
	}.Build()

	// Template: print('{{A}}', '{{B}}')
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('{{A}}', '{{B}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("A")}.Build(),
			}.Build(),
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("B")}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call-pwn")

	// Payload A: \
	// Payload B: , __import__("subprocess").run("echo POwned", shell=True)) #

	req := &ExecutionRequest{
		ToolName: "python-pwn",
		// We use JSON inputs
		ToolInputs: []byte(`{
			"A": "\\",
			"B": ", __import__(\"subprocess\").run(\"echo POwned\", shell=True)) #"
		}`),
	}

	result, err := tool.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed with unexpected error: %v", err)
	}

	resMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not map: %v", result)
	}

	stdout, _ := resMap["stdout"].(string)
	t.Logf("Stdout: %s", stdout)

	// We expect stdout to contain the LITERAL strings:
	// \
	// , __import__...
	// If RCE happened, it would execute.

	// If RCE prevented: stdout should contain "subprocess" (literal string)
	// If RCE executed: stdout should contain output of run (CompletedProcess) AND maybe "POwned"

	// But "POwned" is inside the literal string too!
	// So checking for "POwned" is not enough to distinguish.

	// Check for "__import__" - this is part of code.
	// If executed, output of print('string', result_of_run).
	// result_of_run is CompletedProcess object.
	// It does NOT contain "__import__".
	// So if "__import__" is present, it was NOT executed (it remained a string).

	if strings.Contains(stdout, "__import__") {
		t.Logf("Success: Code was treated as string literal.")
	} else {
		t.Errorf("FAIL: Code execution suspected! Literal '__import__' not found in stdout.")
	}
}
