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

func TestLocalCommandTool_AwkParserBypass_Repro(t *testing.T) {
	// Define a tool that uses 'awk'.
	tool := v1.Tool_builder{
		Name: proto.String("awk-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `awk '{{script}}'` (Quoted context)
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload that attempts to confuse the quote parser.
	// The scanner sees `'` and enters "single quoted string" mode, ignoring the `@`.
	// Awk sees `/'/` as a regex, then the `@` indirect call.
	// BEGIN { /'/; @f("system", "id"); /'/ }
	// We use "sy" "stem" to also bypass keyword check if necessary, though @f uses indirect.
	payload := `BEGIN { /'/; f="sy" "stem"; @f("id"); /'/ }`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// In a vulnerable system (where parser treats ' as quote for awk), this will SUCCEED.
	_, err := localTool.Execute(context.Background(), req)

	if err == nil {
		t.Logf("Vulnerability Reproduced: Awk parser bypass payload %q was allowed!", payload)
		t.Fail()
	} else {
		t.Logf("Blocked: %v", err)
	}
}
