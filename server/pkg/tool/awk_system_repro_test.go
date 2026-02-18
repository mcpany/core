// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_AwkSystemInjection_Repro(t *testing.T) {
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

	// Payload uses gawk's indirect function call feature to call 'system("id")'
	// without using the literal "system" keyword.
	// f = "sy" "stem" -> concatenates to "system"
	// @f("id") -> calls function named "system" with argument "id"
	// This bypasses checkUnquotedKeywords (which checks "sy" and "stem" separately)
	// and checkInterpreterFunctionCalls (which checks for "system(")
    // Note: "sys" is blocked because it is a dangerous keyword for Python, so we split as "sy" + "stem"
	payload := `BEGIN { f="sy" "stem"; @f("id") }`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute the tool.
	// We expect this to fail due to security checks blocking '@'.
	_, err := localTool.Execute(context.Background(), req)

	// If err is nil, it means the security check passed -> Vulnerable.
	if err == nil {
		t.Fatalf("Vulnerability Reproduced: Awk system injection payload %q was allowed!", payload)
	} else {
		t.Logf("Blocked (Expected): %v", err)
	}
}
