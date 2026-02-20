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

	// Payload:
	// We attempt to execute a system command using gawk's indirect function call syntax (@func()).
	// We also split the "system" keyword to evade keyword filtering.
	//
	// BEGIN { f="sy" "stem"; @f("id") }
	//
	// We use "sy" and "stem" because "sys" is blocked (it's a python module keyword).
	//
	// This relies on:
	// 1. Quoted context (args: ["'{{script}}'"]) which skips strict unquoted checks.
	// 2. checkInterpreterFunctionCalls checks "system(" but fails on "sy" "stem".
	// 3. checkUnquotedKeywords checks quotes as delimiters, so it sees "sy", "stem".
	// 4. checkAwkInjection only blocks |, >, <, getline. It allows @ if not fixed.

	payload := `BEGIN { f="sy" "stem"; @f("id") }`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// In a secure system, this Execute call MUST fail.

	_, err := localTool.Execute(context.Background(), req)

	if err == nil {
		t.Fatalf("Vulnerability Detected: Awk system injection payload %q was allowed!", payload)
	}

	// Verify that the error is due to our fix
	if !strings.Contains(err.Error(), "awk injection detected") && !strings.Contains(err.Error(), "potential indirect function call") {
		t.Errorf("Unexpected error message (might be blocked by another check?): %v", err)
	} else {
		t.Logf("Blocked as expected: %v", err)
	}
}
