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

func TestLocalCommandTool_AwkInjection_Repro(t *testing.T) {
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

	// We attempt to pass an awk script that executes a shell command via pipe
	// BEGIN { print "hello" | "sh" }
	// This does not contain single quotes, backticks, or "system(".
	// It relies on awk's pipe functionality to execute commands.

	// We use "cat" instead of "sh" to be safe but demonstrate the pipe is allowed.
	// If the pipe character '|' was blocked, this would fail.
	payload := `BEGIN { print "pwned" | "cat" }`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// In a vulnerable system, this Execute call will SUCCEED because checkInterpreterInjection
	// for single quotes only checks for system(), exec(), etc., and backticks.
	// It misses the pipe operator '|'.

	// However, since we are asserting SECURITY, we want this to FAIL (return error).
	// If it returns nil error, it means the injection passed the filter -> Vulnerability Confirmed.

	_, err := localTool.Execute(context.Background(), req)

	// If err is nil, the protection failed.
	if err == nil {
		t.Logf("Vulnerability Reproduced: Awk injection payload %q was allowed!", payload)
		t.Fail()
	} else {
		t.Logf("Blocked: %v", err)
	}
}
