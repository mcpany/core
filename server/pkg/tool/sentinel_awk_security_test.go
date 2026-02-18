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

	// We attempt to bypass 'system' checks by splitting the keyword.
	// In gawk, indirect function calls allow calling a function stored in a variable.
	// f = "s" "ystem" -> f is "system"
	// @f("id") calls system("id")
	// This payload avoids checkUnquotedKeywords because "system" is split.
	// It avoids checkInterpreterFunctionCalls because "system(" is not present.
	// It avoids checkAwkInjection because '@' is not blocked (yet).
    // We split as "s" "ystem" to avoid triggering on "sys" keyword which is also blocked.
	payload := `BEGIN { f="s" "ystem"; @f("id") }`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// In a vulnerable system, this Execute call will SUCCEED (return nil error).
	// We assert that it FAILS if our security check works.
	// But since we are reproducing the vulnerability, we expect nil error (passed checks).

	_, err := localTool.Execute(context.Background(), req)

	if err == nil {
		t.Logf("Vulnerability Reproduced: Awk system injection payload %q was allowed!", payload)
		t.Fail()
	} else {
		t.Logf("Blocked: %v", err)
	}
}

func TestLocalCommandTool_AwkEmail_Allowed(t *testing.T) {
	// Define a tool that uses 'awk'.
	tool := v1.Tool_builder{
		Name: proto.String("awk-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	// Case 1: Unquoted in YAML (quoteLevel 0)
	// args: ["-v", "var={{email}}"]
	// This usually means user trusts input or input is safe chars.
	// But email contains '@'.
	// checkAwkInjection is called with quoteLevel 0.
	// inQuote = false.
	// stripAwkStrings("email@example.com", false) -> "email@example.com".
	// '@' preceded by 'l'. 'l' is word char.
	// So it should PASS.
	t.Run("Unquoted", func(t *testing.T) {
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"-v", "var={{email}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("email")}.Build()}.Build(),
			},
		}.Build()

		localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-1")

		req := &ExecutionRequest{
			ToolName: "awk-tool",
			Arguments: map[string]interface{}{
				"email": "email@example.com",
			},
		}
		req.ToolInputs, _ = json.Marshal(req.Arguments)

		_, err := localTool.Execute(context.Background(), req)
		if err != nil {
			t.Errorf("Expected success for email@example.com (unquoted), but got error: %v", err)
		}
	})

	// Case 2: Single Quoted in YAML (quoteLevel 2)
	// args: ["-v", "var='{{email}}'"]
	// quoteLevel 2. inQuote = false.
	// stripAwkStrings("email@example.com", false) -> "email@example.com".
	// '@' preceded by 'l'.
	// So it should PASS.
	t.Run("SingleQuoted", func(t *testing.T) {
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"-v", "var='{{email}}'"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("email")}.Build()}.Build(),
			},
		}.Build()

		localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-2")

		req := &ExecutionRequest{
			ToolName: "awk-tool",
			Arguments: map[string]interface{}{
				"email": "email@example.com",
			},
		}
		req.ToolInputs, _ = json.Marshal(req.Arguments)

		_, err := localTool.Execute(context.Background(), req)
		if err != nil {
			t.Errorf("Expected success for email@example.com (single quoted), but got error: %v", err)
		}
	})

	// Case 3: Double Quoted in YAML (quoteLevel 1)
	// args: ["-v", "var=\"{{email}}\""]
	// quoteLevel 1. inQuote = TRUE.
	// stripAwkStrings("email@example.com", true) -> "".
	// No '@'.
	// So it should PASS.
	t.Run("DoubleQuoted", func(t *testing.T) {
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"-v", "var=\"{{email}}\""},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("email")}.Build()}.Build(),
			},
		}.Build()

		localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-3")

		req := &ExecutionRequest{
			ToolName: "awk-tool",
			Arguments: map[string]interface{}{
				"email": "email@example.com",
			},
		}
		req.ToolInputs, _ = json.Marshal(req.Arguments)

		_, err := localTool.Execute(context.Background(), req)
		if err != nil {
			t.Errorf("Expected success for email@example.com (double quoted), but got error: %v", err)
		}
	})
}

func TestLocalCommandTool_AwkSystem_FalsePositive(t *testing.T) {
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

	// Input "filesystem" contains "system".
	// It should be ALLOWED because it's just a string/word, not a function call.
	// But our current broad check blocks it.
	payload := `filesystem`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	if err != nil {
		t.Logf("False Positive Detected: 'filesystem' was blocked: %v", err)
		t.Fail()
	}
}
