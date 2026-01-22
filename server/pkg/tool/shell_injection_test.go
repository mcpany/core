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

	// Case 3: Escaped Single Quote Bypass (Regression Test)
	t.Run("escaped_single_quote_bypass", func(t *testing.T) {
		cmd := "bash"
		// Template: 'foo\'{{input}}'
		// Should be interpreted as: 'foo\' (string) then {{input}} (unquoted)
		args := []string{"-c", "'foo\\'{{input}}'"}
		tool := createTestCommandToolWithArgs(cmd, args)

		// Attempt to inject ;
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "; echo 'pwned'"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected", "escaped single quote should not mask unquoted context")
	})
}

func createTestCommandToolWithArgs(command string, args []string) Tool {
	toolDef := &v1.Tool{Name: proto.String("test-tool")}
	service := &configv1.CommandLineUpstreamService{
		Command: &command,
	}
	callDef := &configv1.CommandLineCallDefinition{
		Args: args,
		Parameters: []*configv1.CommandLineParameterMapping{
			{
				Schema: &configv1.ParameterSchema{Name: proto.String("input")},
			},
		},
	}
	return NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
}

func createTestCommandTool(command string) Tool {
	toolDef := &v1.Tool{Name: proto.String("test-tool")}
	service := &configv1.CommandLineUpstreamService{
		Command: &command,
	}
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-c", "print('{{input}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			{
				Schema: &configv1.ParameterSchema{Name: proto.String("input")},
			},
		},
	}
	return NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
}
