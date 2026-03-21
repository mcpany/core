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

func TestShellInjection_Regression(t *testing.T) {
	// Case 1: python3 (Protected)
	t.Run("python3_protected", func(t *testing.T) {
		cmd := "python3"
		tool := createTestCommandTool(cmd)
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "'; echo 'pwned'; '"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected", "python3 should be protected")
	})

	// Case 2: python3.10 (Protected)
	// This test ensures versioned python binaries are also recognized as shell commands.
	t.Run("python3.10_should_be_protected", func(t *testing.T) {
		cmd := "python3.10"
		tool := createTestCommandTool(cmd)
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "'; echo 'pwned'; '"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected", "python3.10 should be protected")
	})

	// Case 3: mawk (Protected)
	// mawk was previously missing from the protection list.
	t.Run("mawk_protected", func(t *testing.T) {
		cmd := "mawk"
		// We need to use Unquoted input that triggers the check.
		// Custom tool definition for unquoted input
		toolDef := v1.Tool_builder{Name: proto.String("test-tool")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"{{input}}"}, // Unquoted!
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
				}.Build(),
			},
		}.Build()
		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

		req := &ExecutionRequest{
			ToolName: "test",
			// input contains {, }, " which are in dangerousChars list
			ToolInputs: []byte(`{"input": "BEGIN { print \"pwned\" }"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected", "mawk should be protected")
	})
}

func createTestCommandTool(command string) Tool {
	toolDef := v1.Tool_builder{Name: proto.String("test-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &command,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('{{input}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()
	return NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
}
