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

// TestLocalCommandTool_AwkSystemInjection_Security tests that gawk indirect function calls are blocked.
func TestLocalCommandTool_AwkSystemInjection_Security(t *testing.T) {
	// Define a tool that uses 'awk'.
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

	// Payload uses gawk's indirect function call feature to bypass system keyword check.
	// f="s" "ystem"; @f("id")
	// This splits "system" keyword.
	// Previous attempt with "sys" "tem" failed because "sys" is in dangerousKeywords list (for Python).
	// "s" and "ystem" are not in the list.
	payload := `BEGIN { f="s" "ystem"; @f("id") }`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// We expect this Execute call to FAIL (return an error).
	// If it succeeds, the vulnerability is present.

	_, err := localTool.Execute(context.Background(), req)

	if err == nil {
		t.Logf("Vulnerability Reproduced: Awk system injection payload %q was allowed!", payload)
		t.Fail()
	} else {
		t.Logf("Blocked: %v", err)
	}
}

// TestLocalCommandTool_AwkStringWithAtSymbol tests that legitimate use of '@' in strings is allowed.
func TestLocalCommandTool_AwkStringWithAtSymbol(t *testing.T) {
	// Define a tool that uses 'awk'.
	tool := v1.Tool_builder{
		Name: proto.String("awk-tool-safe"),
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

	// Payload uses '@' inside double quotes, which should be safe.
	// We use 'print' so it's a valid awk script.
	payload := `BEGIN { print "email@example.com" }`

	req := &ExecutionRequest{
		ToolName: "awk-tool-safe",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// We expect this Execute call to NOT return a security error.
	_, err := localTool.Execute(context.Background(), req)

	if err != nil {
		if strings.Contains(err.Error(), "awk injection detected") {
			t.Errorf("False Positive: Safe payload %q was blocked: %v", payload, err)
		} else {
			// Other errors are acceptable (e.g. awk not found)
			t.Logf("Execution failed (expected/acceptable): %v", err)
		}
	} else {
		t.Logf("Success: Safe payload allowed")
	}
}
