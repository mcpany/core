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
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_RubyPercentX_Injection(t *testing.T) {
	// This test demonstrates that Ruby %x execution is BLOCKED
	// inside Unquoted arguments passed to ruby -e.

	t.Parallel()

	tool := v1.Tool_builder{
		Name: proto.String("ruby-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()

	// Ruby script argument WITHOUT quotes.
    // This simulates: ruby -e {{input}}
    // Or: bash -c "ruby -e {{input}}" (where bash strips quotes)
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{input}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, []*configv1.CallPolicy{}, "call-id")

	// Payload: %x(echo injected)
	// We don't need 'p' prefix, just execution.
	payload := "%x(echo injected)"

	req := &ExecutionRequest{
		ToolName: "ruby-tool",
		Arguments: map[string]interface{}{
			"input": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	// We expect this to be blocked by our hardening.
	if err != nil {
		// If it errors, check if it's the expected injection error
		if strings.Contains(err.Error(), "injection detected") {
			t.Logf("Successfully blocked: %v", err)
		} else {
			t.Logf("Blocked with unexpected error: %v", err)
		}
	} else {
		t.Error("Vulnerability Reproduced: percent-x execution passed validation!")
	}
    assert.Error(t, err)
}

func TestLocalCommandTool_RubyPercentX_NestedShell_Injection(t *testing.T) {
	// This test demonstrates that Ruby %x execution is BLOCKED
	// when nested inside a shell command, even if quotes appear in the template.
	// bash -c "ruby -e '{{input}}'"

	t.Parallel()

	tool := v1.Tool_builder{
		Name: proto.String("bash-ruby-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("bash"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "ruby -e '{{input}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, []*configv1.CallPolicy{}, "call-id")

	// Payload: %x(echo injected)
	// Even though it is inside single quotes in the template, bash removes them.
	// Ruby receives Unquoted input.
	payload := "%x(echo injected)"

	req := &ExecutionRequest{
		ToolName: "bash-ruby-tool",
		Arguments: map[string]interface{}{
			"input": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	if err != nil {
		if strings.Contains(err.Error(), "injection detected") {
			t.Logf("Successfully blocked: %v", err)
		} else {
			t.Logf("Blocked with unexpected error: %v", err)
		}
	} else {
		t.Error("Vulnerability Reproduced: Nested Shell percent-x execution passed validation!")
	}
	assert.Error(t, err)
}
