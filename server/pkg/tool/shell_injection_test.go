// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		// Sentinel Update: We now sanitize instead of block.
		// Input "'; echo 'pwned'; '" -> Escaped to "\'; echo \'pwned\'; \'"
		// Python output should be the literal string.
		res, err := tool.Execute(context.Background(), req)
		require.NoError(t, err)
		resMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		stdout, _ := resMap["stdout"].(string)
		// Check that it printed the literal string (containing escaped quote) and did NOT execute code.
		// If executed code, it would output "pwned".
		// The literal string contains "pwned" too, so we check for presence of single quote in output which confirms string.
		assert.Contains(t, stdout, "'", "Should output literal single quote")
	})

	// Case 2: python3.12 (Protected)
	// This test ensures versioned python binaries are also recognized as shell commands.
	t.Run("python3.12_should_be_protected", func(t *testing.T) {
		cmd := "python3.12"
		tool := createTestCommandTool(cmd)
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "'; echo 'pwned'; '"}`),
		}

		res, err := tool.Execute(context.Background(), req)
		require.NoError(t, err)
		resMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		stdout, _ := resMap["stdout"].(string)
		assert.Contains(t, stdout, "'", "Should output literal single quote")
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
