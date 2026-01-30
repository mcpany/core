// Copyright 2025 Author(s) of MCP Any
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

func TestCommandInjection_WrapperBypass(t *testing.T) {
	// Case: Wrapper bypass attempt using 'timeout'
	// 'timeout' is now in the isShellCommand list, so it should enforce shell injection checks.
	t.Run("timeout_wrapper_bypass", func(t *testing.T) {
		cmd := "timeout"
		// Configuration: timeout 1s sh -c {{input}}
		// If 'timeout' is detected as a shell, {{input}} will be checked for shell injection.
		toolDef := v1.Tool_builder{Name: proto.String("test-tool")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}.Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"1s", "sh", "-c", "{{input}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
				}.Build(),
			},
		}.Build()

		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

		// Input contains dangerous characters that should be blocked
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "echo safe; echo pwned"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		// We expect an error because 'timeout' is now flagged as a shell.
		assert.Error(t, err, "Security fix check: Should return error")
		if err != nil {
			assert.Contains(t, err.Error(), "shell injection detected")
		}
	})
}
