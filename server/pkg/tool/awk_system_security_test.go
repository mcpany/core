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

func TestLocalCommandTool_AwkSystemInjection_Security(t *testing.T) {
	// Define a tool that uses 'gawk' (which supports indirect function calls).
	// We use 'gawk' as the base command to ensure isAwk check passes and gawk features are relevant.
	tool := v1.Tool_builder{
		Name: proto.String("awk-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("gawk"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `gawk '{{script}}'`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	// Payload uses gawk indirect function call: @var(args)
	// We construct "system" dynamically to bypass checkUnquotedKeywords and checkInterpreterFunctionCalls.
	// f = "s" "ystem"; @f("id")
	// "sys" was blocked because it's a python module keyword.
	// "s" and "ystem" are safe keywords.
	payload := `BEGIN { f="s" "ystem"; @f("id") }`

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute should block this if secure. If vulnerable, it returns nil error (or fails execution but passes security check).
	_, err := localTool.Execute(context.Background(), req)

	if err == nil {
		t.Errorf("Vulnerability Confirmed: Payload %q was allowed!", payload)
	} else {
		t.Logf("Blocked: %v", err)
	}

	// Test valid usage: quoted string with @
	validPayload := `BEGIN { print "contact@example.com" }`
	reqValid := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": validPayload,
		},
	}
	reqValid.ToolInputs, _ = json.Marshal(reqValid.Arguments)

	_, err = localTool.Execute(context.Background(), reqValid)
	if err != nil {
		t.Errorf("Unexpectedly blocked valid payload %q: %v", validPayload, err)
	}
}
