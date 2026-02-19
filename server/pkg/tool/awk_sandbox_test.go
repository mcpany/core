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

func TestLocalCommandTool_AwkSandbox(t *testing.T) {
	// Define a tool that uses 'awk'.
	tool := v1.Tool_builder{
		Name: proto.String("awk-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	t.Run("Usability_Email_Allowed", func(t *testing.T) {
		// Use unquoted args to test static filters
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"{{script}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
			},
		}.Build()

		localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

		// Payload containing '@' but no other dangerous chars
		// This should PASS static checks (checkAwkInjection)
		payload := `user@example.com`

		req := &ExecutionRequest{
			ToolName: "awk-tool",
			Arguments: map[string]interface{}{
				"script": payload,
			},
		}
		req.ToolInputs, _ = json.Marshal(req.Arguments)

		_, err := localTool.Execute(context.Background(), req)

		// We expect NO security error.
		// It might fail with awk syntax error or timeout, which is fine.
		if err != nil && strings.Contains(err.Error(), "injection detected") {
			t.Fatalf("Usability regression: '@' was blocked by static analysis: %v", err)
		}
	})

	t.Run("Security_RCE_Mitigated", func(t *testing.T) {
		// Use quoted args to trigger quoted context checks and attempt bypass
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"'{{script}}'"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
			},
		}.Build()

		localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

		// Payload attempting to use @f indirect call.
		// Even if static check allows it (due to usability fix),
		// the Sandbox should prevent RCE or awk should fail syntax error.
		payload := `BEGIN { f="s" "ystem"; @f("echo pwned") }`

		req := &ExecutionRequest{
			ToolName: "awk-tool",
			Arguments: map[string]interface{}{
				"script": payload,
			},
		}
		req.ToolInputs, _ = json.Marshal(req.Arguments)

		result, err := localTool.Execute(context.Background(), req)

		if err != nil {
			// If blocked by static check, that's also fine (secure).
			t.Logf("Blocked by execution error: %v", err)
			return
		}

		resMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Unexpected result type: %T", result)
		}
		stdout := resMap["stdout"].(string)

		if strings.Contains(stdout, "pwned") {
			t.Fatalf("Vulnerability! Awk system injection executed: %s", stdout)
		}
	})
}
