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

func TestShellInjection_Repro(t *testing.T) {
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

	// Case 2: python3.10 (Vulnerable - Bug Repro)
	// This test fails if python3.10 is NOT recognized as a shell command.
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
