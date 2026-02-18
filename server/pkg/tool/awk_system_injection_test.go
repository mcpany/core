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
	// The single quotes around {{script}} put us in quoteLevel=2 context.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload that bypasses keyword check by splitting "system"
	// and uses indirect function call via '@' (gawk feature).
	// BEGIN { f="sy" "stem"; @f("id") }
	// We use "sy" "stem" because "sys" is blocked as a keyword (for Python).
	payload := `BEGIN { f="sy" "stem"; @f("id") }`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// In a hardened system, this Execute call will FAIL (return error)
	// because checkInterpreterFunctionCalls now detects the '@' character in unquoted context.

	_, err := localTool.Execute(context.Background(), req)

	// If err is nil, the protection failed.
	if err == nil {
		t.Logf("Vulnerability: Awk system injection payload %q was allowed!", payload)
		t.Fail()
	} else {
		t.Logf("Correctly blocked: %v", err)
	}
}

func TestLocalCommandTool_AwkSystemKeyword_Blocked(t *testing.T) {
	// Verify that direct usage of "system" is blocked (sanity check).
	tool := v1.Tool_builder{
		Name: proto.String("awk-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	payload := `BEGIN { system("id") }`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	if err == nil {
		t.Errorf("Security check failed: Direct usage of 'system' was allowed!")
	} else {
		t.Logf("Correctly blocked direct usage: %v", err)
	}
}

func TestLocalCommandTool_AwkValidAt_Allowed(t *testing.T) {
	// Verify that valid usage of '@' in strings/regex is allowed.
	tool := v1.Tool_builder{
		Name: proto.String("awk-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload with '@' inside double quotes (string).
	payload := `BEGIN { print "contact@me.com" }`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// Since "awk" execution might fail if we don't provide input or if environment is restricted,
	// we mainly check that the SECURITY error is NOT returned.
	// If it fails with "exit status 1" or similar, it means it PASSED the security check.
	// The security check returns specific error messages.

	if err != nil {
		errStr := err.Error()
		// If the error message contains the security violation message, it's a false positive.
		// Note: The error message might be wrapped in "parameter 'script': ..."
		if errStr == "parameter \"script\": awk injection detected: value contains '@' (potential indirect function call)" {
			t.Errorf("False positive: Valid usage of '@' was blocked: %v", err)
		} else if errStr == "awk injection detected: value contains '@' (potential indirect function call)" {
			t.Errorf("False positive: Valid usage of '@' was blocked: %v", err)
		} else {
			t.Logf("Execution failed (expected) but security check passed: %v", err)
		}
	} else {
		t.Logf("Allowed as expected")
	}
}

func TestLocalCommandTool_AwkUnquotedAt_Blocked(t *testing.T) {
	// Verify that usage of '@' in unquoted context is blocked.
	tool := v1.Tool_builder{
		Name: proto.String("awk-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `awk {{script}}` (Unquoted)
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload with '@' unquoted.
	payload := `@include "evil.awk"`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	if err == nil {
		t.Errorf("Security check failed: Unquoted usage of '@' was allowed!")
	} else {
		// Expect "shell injection detected: value contains dangerous character '@'"
		// because checkUnquotedInjection is called.
		t.Logf("Correctly blocked unquoted usage: %v", err)
	}
}
