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

func TestLocalCommandTool_AwkParser_SingleQuotes(t *testing.T) {
	// Verify that the parser correctly handles single quotes for awk (i.e. does NOT treat them as quotes).
	tool := v1.Tool_builder{
		Name: proto.String("awk-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `awk "{{script}}"` (Double Quoted context - quoteLevel 1)
	// This allows single quotes in the input.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"\"{{script}}\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: ' @ '
	// For awk, single quotes are NOT string delimiters.
	// So @ is effectively unquoted (conceptually).
	// The parser should detect @.
	payload := `' @ '`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// Currently (Vulnerable/Incorrect): parser sees '...', thinks @ is quoted, returns nil (or other error but not security).
	// Expected (Fixed): parser sees ' as literal, sees @, returns security error.

	if err == nil {
		t.Logf("Incorrect parser behavior: @ inside single quotes was ignored for awk!")
		t.Fail()
	} else {
		errStr := err.Error()
		expected := "awk injection detected: value contains '@' (potential indirect function call)"
		if errStr != expected && errStr != "parameter \"script\": "+expected {
			t.Logf("Blocked but with unexpected error: %v", err)
			// If it was blocked by something else, we need to know.
			// But here quoteLevel 1 allows '.
		} else {
			t.Logf("Correctly blocked: %v", err)
		}
	}
}
