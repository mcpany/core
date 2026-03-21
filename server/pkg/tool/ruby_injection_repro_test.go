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

func TestLocalCommandTool_RubyInjection_Backtick(t *testing.T) {
	// This test demonstrates that Ruby interpolation #{...} is BLOCKED
	// inside backticked arguments.

	t.Parallel()

	tool := v1.Tool_builder{
		Name: proto.String("ruby-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
		Local:   proto.Bool(true),
	}.Build()

	// Ruby script that uses backticks: `echo #{input}`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "`echo {{input}}`"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: #{system('echo injected')}
	payload := "#{system('echo injected')}"

	req := &ExecutionRequest{
		ToolName: "ruby-tool",
		Arguments: map[string]interface{}{
			"input": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "ruby interpolation injection detected"), "Expected ruby interpolation error, got: %v", err)
}

func TestLocalCommandTool_PHPInjection_Backtick(t *testing.T) {
	// This test demonstrates that PHP interpolation $... is BLOCKED
	// inside backticked arguments.

	t.Parallel()

	tool := v1.Tool_builder{
		Name: proto.String("php-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("php"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-r", "`echo {{input}}`;"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: ${system('echo injected')}
	payload := "${system('echo injected')}"

	req := &ExecutionRequest{
		ToolName: "php-tool",
		Arguments: map[string]interface{}{
			"input": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "variable interpolation injection detected"), "Expected variable interpolation error, got: %v", err)
}
