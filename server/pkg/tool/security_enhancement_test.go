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

func TestSecurityEnhancements(t *testing.T) {
	// Case 1: Rsync Shell Injection
	// rsync is now in the blacklist, so it should trigger strict checking
	t.Run("rsync_shell_injection", func(t *testing.T) {
		cmd := "rsync"
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "foo; rm -rf /"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected")
	})

	// Case 2: Dangerous Environment Variable
	t.Run("dangerous_env_var", func(t *testing.T) {
		cmd := "ls"
		toolDef := &v1.Tool{Name: proto.String("test-tool")}
		service := &configv1.CommandLineUpstreamService{
			Command: &cmd,
		}
		callDef := &configv1.CommandLineCallDefinition{
			Parameters: []*configv1.CommandLineParameterMapping{
				{
					Schema: &configv1.ParameterSchema{Name: proto.String("PYTHONPATH")},
				},
			},
		}
		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"PYTHONPATH": "/tmp/malicious"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dangerous environment variable name")
	})

	// Case 3: Dangerous Environment Variable (case insensitive)
	t.Run("dangerous_env_var_case", func(t *testing.T) {
		cmd := "ls"
		toolDef := &v1.Tool{Name: proto.String("test-tool")}
		service := &configv1.CommandLineUpstreamService{
			Command: &cmd,
		}
		callDef := &configv1.CommandLineCallDefinition{
			Parameters: []*configv1.CommandLineParameterMapping{
				{
					Schema: &configv1.ParameterSchema{Name: proto.String("pythonpath")},
				},
			},
		}
		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"pythonpath": "/tmp/malicious"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dangerous environment variable name")
	})
}
