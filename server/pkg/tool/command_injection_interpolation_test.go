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

func TestCommandInjection_RubyInterpolation(t *testing.T) {
	// Ruby string interpolation #{...} works in double-quoted strings
	t.Run("ruby_double_quoted_interpolation", func(t *testing.T) {
		cmd := "ruby"
		// Template: puts "{{input}}"
		// We pass it as: ruby -e 'puts "{{input}}"'
		// Wait, if we use single quotes for the shell arg, then the inner quotes matter.
		// LocalCommandTool uses exec.Command, so no outer shell quotes.
		// Args: ["-e", "puts \"{{input}}\""]

		tool := createTestCommandToolWithArgs(cmd, []string{"-e", "puts \"{{input}}\""})

		// Input attempts to run system command
		// #{system('echo pwned')}
		// We expect this to be BLOCKED by the security check
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "#{system('echo pwned')}"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		if err == nil {
			t.Fatal("Expected error due to ruby interpolation injection, got nil")
		}
		assert.Contains(t, err.Error(), "shell injection detected")
	})
}

func TestCommandInjection_PythonFString(t *testing.T) {
	// Python f-strings f'{...}' evaluate expressions
	t.Run("python_fstring_interpolation", func(t *testing.T) {
		cmd := "python3"
		// Args: ["-c", "print(f'{{input}}')"]

		tool := createTestCommandToolWithArgs(cmd, []string{"-c", "print(f'{{input}}')"})

		// Input attempts to run system command
		// {__import__("os").system("echo pwned")}
		// We use double quotes to avoid triggering the single-quote check
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "{__import__(\"os\").system(\"echo pwned\")}"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		if err == nil {
			t.Fatal("Expected error due to python f-string injection, got nil")
		}
		assert.Contains(t, err.Error(), "shell injection detected")
	})

	// Python raw f-strings rf'{...}' evaluate expressions
	t.Run("python_raw_fstring_interpolation", func(t *testing.T) {
		cmd := "python3"
		// Args: ["-c", "print(rf'{{input}}')"]

		tool := createTestCommandToolWithArgs(cmd, []string{"-c", "print(rf'{{input}}')"})

		// Input attempts to run system command
		// {__import__("os").system("echo pwned")}
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "{__import__(\"os\").system(\"echo pwned\")}"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		if err == nil {
			t.Fatal("Expected error due to python raw f-string injection, got nil")
		}
		assert.Contains(t, err.Error(), "shell injection detected")
	})
}

func createTestCommandToolWithArgs(command string, args []string) Tool {
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
