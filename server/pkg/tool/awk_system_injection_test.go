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

func TestLocalCommandTool_AwkSystem_Injection(t *testing.T) {
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

	// Payload: Use gawk's indirect function call feature to bypass checkInterpreterFunctionCalls
	// f="sy" "stem"; @f("id")
	// This splits "system" keyword, so checkUnquotedKeywords won't match it.
	// And since quotes are used, checkInterpreterFunctionCalls runs on normalized string:
	// "BEGIN{f="sy""stem";@f("id")}"
	// This does NOT contain "system(", so checkInterpreterFunctionCalls won't match.
	// And checkAwkInjection only checks |, >, <, getline.

	// We use "sy" and "stem" to avoid "sys" keyword which is blocked as dangerous keyword (Python).
	payload := `BEGIN { f="sy" "stem"; @f("id") }`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// If err is nil, the protection failed.
	if err == nil {
		t.Logf("Vulnerability Reproduced: Awk system injection payload %q was allowed!", payload)
		// We expect this to fail (security check should fail).
		// But in repro mode, failure means vulnerability exists.
		// So we fail the test if it passes security checks.
		t.Fail()
	} else {
		t.Logf("Blocked: %v", err)
	}
}
