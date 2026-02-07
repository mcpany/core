// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_RubyRCE_ExecNoParen(t *testing.T) {
	t.Parallel()

	// Ruby tool running a script from an argument (single quoted)
	tool := v1.Tool_builder{
		Name: proto.String("ruby-rce"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: exec "id"
	// This avoids single quotes, backticks, and "system(", "exec(".
	// In Ruby, `exec "id"` is valid syntax.
	payload := `exec "id"`

	req := &ExecutionRequest{
		ToolName: "ruby-rce",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// We expect this to execute safely if our mitigations are working.
	// But if it's vulnerable, it will execute without error (mock executor handles it).
	// We want to assert that the SECURITY CHECK blocks it.

	_, err := localTool.Execute(context.Background(), req)

	// Vulnerability Confirmation:
	// If err is nil, it means the payload was allowed through.
	if err == nil {
		t.Logf("VULNERABILITY CONFIRMED: Allowed payload %q", payload)
		t.Fail()
	} else {
		// If it was blocked, we expect a security error
		t.Logf("Blocked with error: %v", err)
		assert.Contains(t, err.Error(), "injection detected")
	}
}

func TestLocalCommandTool_RubyRCE_PercentX(t *testing.T) {
	t.Parallel()

	tool := v1.Tool_builder{
		Name: proto.String("ruby-rce-percent"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: %x(id)
	// This executes `id` command in Ruby.
	payload := `%x(id)`

	req := &ExecutionRequest{
		ToolName: "ruby-rce-percent",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	if err == nil {
		t.Logf("VULNERABILITY CONFIRMED: Allowed payload %q", payload)
		t.Fail()
	} else {
		t.Logf("Blocked with error: %v", err)
		assert.Contains(t, err.Error(), "injection detected")
	}
}
