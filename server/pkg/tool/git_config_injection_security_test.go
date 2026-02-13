// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Git_Config_Injection_Security(t *testing.T) {
	// This test attempts to demonstrate that git allows command execution via -c configuration.
	// Specifically core.pager or aliases.

	toolProto := v1.Tool_builder{
		Name: proto.String("git-pager-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
	}.Build()

	// git -c alias.pwn={{val}} pwn
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "alias.pwn={{val}}", "pwn"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("val"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// The payload: !echo PWNED
	payload := "!echo PWNED"
	// We need to escape quotes for JSON
	inputs := fmt.Sprintf(`{"val": "%s"}`, payload)

	req := &ExecutionRequest{
		ToolName:   "git-pager-tool",
		ToolInputs: []byte(inputs),
	}

	_, err := tool.Execute(context.Background(), req)
	if err == nil {
		t.Errorf("Expected security error for alias injection, got nil")
	} else if !strings.Contains(err.Error(), "git config values") {
		t.Errorf("Expected git config injection error for alias, got: %v", err)
	}

	// Test 2: Split argument injection (core.pager)
	// git -c core.pager={{pager}} log
	callDef2 := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "core.pager={{pager}}", "log"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("pager"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool2 := NewLocalCommandTool(toolProto, service, callDef2, nil, "call-id-2")
	// Payload: sh -c id (spaces are dangerous here)
	inputs2 := `{"pager": "sh -c id"}`
	req2 := &ExecutionRequest{
		ToolName:   "git-pager-tool-2",
		ToolInputs: []byte(inputs2),
	}

	_, err2 := tool2.Execute(context.Background(), req2)
	if err2 == nil {
		t.Errorf("Expected security error for pager injection, got nil")
	} else if !strings.Contains(err2.Error(), "git config values") {
		t.Errorf("Expected git config injection error for pager, got: %v", err2)
	}
}
