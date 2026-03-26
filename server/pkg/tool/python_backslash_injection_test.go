// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtrBackslash(s string) *string { return &s }

func TestNodeBackslashInjection(t *testing.T) {
	// Vulnerability: Injecting backslash into a single-quoted string in an interpreter context (Node.js)
	// allows escaping the closing quote and continuing the string until the next quote,
	// potentially treating subsequent code/parameters as string content and exposing intervening code.
	//
	// Command: node -e "console.log('{{A}}'); console.log('{{B}}')"
	//
	// Attack: A="\", B="process.exit(42); //"
	// Arg: console.log('\'); console.log('process.exit(42); //')
	//
	// Parsed as:
	// console.log(' ... console.log(');
	// process.exit(42); // ...

	service := configv1.CommandLineUpstreamService_builder{
		Command: strPtrBackslash("node"),
	}.Build()

	// Args: node -e "console.log('{{A}}'); console.log('{{B}}')"
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "console.log('{{A}}'); console.log('{{B}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: strPtrBackslash("A"),
				}.Build(),
			}.Build(),
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: strPtrBackslash("B"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := pb.Tool_builder{
		Name: strPtrBackslash("node_backslash"),
	}.Build()

	policies := []*configv1.CallPolicy{}
	tool := NewLocalCommandTool(toolProto, service, callDef, policies, "call-id")

	// Payload: A ends with backslash. B contains the payload.
	// Add ); to B to close the function call (console.log) and then execute payload.
	inputs := map[string]interface{}{
		"A": "\\",
		"B": "); process.exit(42); //",
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "node_backslash",
		ToolInputs: inputBytes,
	}

	result, err := tool.Execute(context.Background(), req)

	if err != nil {
		// If it was blocked by security check, err will be non-nil.
		t.Logf("Execution blocked as expected: %v", err)
		assert.Contains(t, err.Error(), "injection detected")
	} else {
		// If it ran, check return code.
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok)

		var returnCode int
		if rc, ok := resMap["return_code"].(float64); ok {
			returnCode = int(rc)
		} else if rc, ok := resMap["return_code"].(int); ok {
			returnCode = rc
		}

		t.Logf("Return Code: %d", returnCode)
		stdout, _ := resMap["stdout"].(string)
		stderr, _ := resMap["stderr"].(string)
		t.Logf("Stdout: %s", stdout)
		t.Logf("Stderr: %s", stderr)

		if returnCode == 42 {
			t.Fatal("Vulnerability check failed: Code injection successful (exit code 42)")
		}
		t.Fatal("Vulnerability check failed: Execution proceeded (it should have been blocked)")
	}
}
