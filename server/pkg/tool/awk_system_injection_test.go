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

func TestLocalCommandTool_AwkSystem_Blocked(t *testing.T) {
	// Define a tool that uses 'gawk', which is an interpreter.
	tool := v1.Tool_builder{
		Name: proto.String("gawk-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("gawk"), // Explicitly target gawk
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `gawk '{{script}}'`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload with indirect function call (RCE attempt)
	payload := `BEGIN { f="s" "y" "s" "tem"; @f("id") }`

	req := &ExecutionRequest{
		ToolName: "gawk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute should fail EITHER because checkAwkInjection blocks it (if sandbox not supported)
	// OR because sandbox blocks it (if sandbox supported).
	result, err := localTool.Execute(context.Background(), req)

	if err != nil {
		// Blocked by injection check
		if strings.Contains(err.Error(), "awk injection detected: value contains '@'") {
			t.Logf("Blocked by injection check as expected: %v", err)
		} else {
			t.Logf("Blocked with unexpected error: %v", err)
		}
	} else {
		// Execution attempted. Check if blocked by sandbox.
		resMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Result is not a map")
		}
		stderr, _ := resMap["stderr"].(string)

		if strings.Contains(stderr, "fatal: 'system' function not allowed in sandbox mode") {
			t.Logf("Blocked by sandbox as expected: %s", stderr)
		} else {
			// If it succeeded (no sandbox error), check exit code.
			// gawk returns 0 if execution succeeds.
			// Sandbox violation returns non-zero.
			retCode, _ := resMap["return_code"].(int)
			if retCode == 0 {
				t.Errorf("Vulnerability Confirmed: Awk indirect function call payload %q was allowed (exit code 0)!", payload)
			} else {
				t.Logf("Command failed (good), stderr: %s", stderr)
			}
		}
	}
}
