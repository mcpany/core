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

func TestCommandInjection_Advanced(t *testing.T) {
	// Case 1: Unquoted shell injection
	t.Run("unquoted_shell_injection", func(t *testing.T) {
		cmd := "sh"
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}") // Unquoted
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "foo; rm -rf /"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected")
	})

	// Case 2: Double quoted shell injection
	t.Run("double_quoted_shell_injection", func(t *testing.T) {
		cmd := "sh"
		tool := createTestCommandToolWithTemplate(cmd, "\"{{input}}\"")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "` + "`whoami`" + `"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected")
	})

	// Case 3: Argument injection (leading dash)
	t.Run("argument_injection", func(t *testing.T) {
		cmd := "ls"
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "-la"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "argument injection detected")
	})

	// Case 4: Path traversal
	t.Run("path_traversal", func(t *testing.T) {
		cmd := "cat"
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "../../etc/passwd"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal attempt detected")
	})

    // Case 5: Absolute path
	t.Run("absolute_path", func(t *testing.T) {
		cmd := "cat"
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "/etc/passwd"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "absolute path detected")
	})

	// Case 6: Shell injection bypass attempt with versioned binary (e.g. python-3.14)
	t.Run("versioned_binary_bypass", func(t *testing.T) {
		cmd := "python-3.14" // Should be treated as python
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "print('hello'); import os; os.system('rm -rf /')"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected")
	})

	// Case 7: Improved quote detection allows safe chars in quotes
	t.Run("improved_quote_detection", func(t *testing.T) {
		cmd := "python"
		tool := createTestCommandToolWithTemplate(cmd, "print('Prefix: {{input}}')")
		// This input is safe in python string but blocked by strict check currently
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "foo; bar"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.NoError(t, err)
	})
}

func createTestCommandToolWithTemplate(command string, template string) Tool {
	toolDef := &v1.Tool{Name: proto.String("test-tool")}
	service := &configv1.CommandLineUpstreamService{
		Command: &command,
	}
	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetArgs([]string{"-c", template})

	param := &configv1.CommandLineParameterMapping{}
	schema := &configv1.ParameterSchema{}
	schema.SetName("input")
	param.SetSchema(schema)

	callDef.SetParameters([]*configv1.CommandLineParameterMapping{param})
	return NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
}
