// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
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
}

func createTestCommandToolWithTemplate(command string, template string) Tool {
	toolDef := &v1.Tool{Name: proto.String("test-tool")}
	service := &configv1.CommandLineUpstreamService{
		Command: &command,
	}
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-c", template},
		Parameters: []*configv1.CommandLineParameterMapping{
			{
				Schema: &configv1.ParameterSchema{Name: proto.String("input")},
			},
		},
	}
	return NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
}
