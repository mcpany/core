// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestInterpreterInjection(t *testing.T) {
	// Case 1: Perl Array Interpolation Injection
	// Perl interpolates arrays (@array) in double-quoted strings.
	// Attack: @{[system("echo pwned")]}
	t.Run("perl_array_interpolation", func(t *testing.T) {
		cmd := "perl"
		tool := createInterpreterTestTool(cmd, []string{"-e", "print \"{{input}}\""})

		input := "@{[system('echo pwned')]}"
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "` + input + `"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		if assert.Error(t, err, "Expected security error, but execution proceeded (vulnerability exists)") {
			assert.Contains(t, err.Error(), "shell injection detected", "Perl array interpolation should be blocked")
		}
	})

	// Case 2: Tcl Command Substitution Injection
	// Tcl/Expect interpolates commands ([cmd]) in double-quoted strings.
	t.Run("tcl_command_substitution", func(t *testing.T) {
		cmd := "expect"
		tool := createInterpreterTestTool(cmd, []string{"-c", "puts \"{{input}}\""})

		input := "[exec echo pwned]"
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "` + input + `"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		if assert.Error(t, err, "Expected security error, but execution proceeded (vulnerability exists)") {
			assert.Contains(t, err.Error(), "shell injection detected", "Tcl command substitution should be blocked")
		}
	})

	// Case 3: Ruby Interpolation Injection
	// Ruby interpolates #{...} in double-quoted strings.
	t.Run("ruby_interpolation", func(t *testing.T) {
		cmd := "ruby"
		tool := createInterpreterTestTool(cmd, []string{"-e", "puts \"{{input}}\""})

		input := "#{system('echo pwned')}"
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "` + input + `"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		if assert.Error(t, err, "Expected security error, but execution proceeded (vulnerability exists)") {
			assert.Contains(t, err.Error(), "shell injection detected", "Ruby interpolation should be blocked")
		}
	})
}

func createInterpreterTestTool(command string, args []string) Tool {
	toolDef := v1.Tool_builder{Name: proto.String("test-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &command,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: args,
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()
	return NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
}
