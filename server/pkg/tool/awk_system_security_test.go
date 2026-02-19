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

func TestLocalCommandTool_AwkSystem_Security(t *testing.T) {
	// Define a tool that uses 'awk', which is in isShellCommand list.
	tool := v1.Tool_builder{
		Name: proto.String("awk-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `awk '{{script}}'`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	t.Run("Block Indirect Call", func(t *testing.T) {
		// We attempt to bypass checkInterpreterFunctionCalls by splitting "system".
		// gawk supports indirect function calls via `@var`.
		// BEGIN { f="sy" "stem"; @f("id") }
		payload := `BEGIN { f="sy" "stem"; @f("id") }`

		req := &ExecutionRequest{
			ToolName: "awk-tool",
			Arguments: map[string]interface{}{
				"script": payload,
			},
		}
		req.ToolInputs, _ = json.Marshal(req.Arguments)

		// In a SECURE system, this Execute call MUST FAIL.
		_, err := localTool.Execute(context.Background(), req)

		// If err is nil, the protection failed.
		if err == nil {
			t.Errorf("Vulnerability NOT Fixed: Awk system() injection payload %q was allowed!", payload)
		} else {
			// Verify it was blocked for the right reason
			if !strings.Contains(err.Error(), "awk injection detected: value contains '@'") {
				t.Errorf("Unexpected error message: %v", err)
			} else {
				t.Logf("Successfully blocked: %v", err)
			}
		}
	})

	t.Run("Allow Safe String", func(t *testing.T) {
		// Valid use case: printing an email address (contains @) inside a string.
		// BEGIN { print "contact@example.com" }
		payload := `BEGIN { print "contact@example.com" }`

		req := &ExecutionRequest{
			ToolName: "awk-tool",
			Arguments: map[string]interface{}{
				"script": payload,
			},
		}
		req.ToolInputs, _ = json.Marshal(req.Arguments)

		// This should PASS (no error) because @ is inside quotes.
		// NOTE: This will fail if the system running this test does not have 'awk' installed,
		// because Execute actually tries to run the command.
		// However, the SECURITY CHECK happens BEFORE execution.
		// If it fails with "awk injection detected", then our fix is broken (false positive).
		// If it fails with "exec: ... executable file not found", that means security check passed.

		_, err := localTool.Execute(context.Background(), req)

		if err != nil {
			if strings.Contains(err.Error(), "awk injection detected") {
				t.Errorf("False Positive: Safe payload %q was blocked: %v", payload, err)
			} else {
				// Other errors (like execution failure) are expected if awk is missing/fails,
				// but verify it's not a security block.
				t.Logf("Execution failed (expected if awk missing), but passed security check: %v", err)
			}
		} else {
			t.Logf("Safe payload passed security check and execution.")
		}
	})

    t.Run("Allow Safe Pipe String", func(t *testing.T) {
		// Valid use case: printing a pipe char inside a string.
		// BEGIN { print "a|b" }
		payload := `BEGIN { print "a|b" }`

		req := &ExecutionRequest{
			ToolName: "awk-tool",
			Arguments: map[string]interface{}{
				"script": payload,
			},
		}
		req.ToolInputs, _ = json.Marshal(req.Arguments)

		_, err := localTool.Execute(context.Background(), req)

		if err != nil {
			if strings.Contains(err.Error(), "awk injection detected") {
				t.Errorf("False Positive: Safe payload %q was blocked: %v", payload, err)
			} else {
				t.Logf("Execution failed (expected), but passed security check: %v", err)
			}
		} else {
			t.Logf("Safe payload passed security check and execution.")
		}
	})
}
