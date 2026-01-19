// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestSecurityEnhancements(t *testing.T) {
	t.Run("rsync_shell_injection_prevention", func(t *testing.T) {
		// rsync is now in the blacklist.
		// If we try to inject shell characters, it should be caught.
		cmd := "rsync"
		// Config: rsync -e {{input}} source dest
		template := "{{input}}"

		toolDef := &v1.Tool{Name: proto.String("test-rsync")}
		service := &configv1.CommandLineUpstreamService{
			Command: &cmd,
		}
		callDef := &configv1.CommandLineCallDefinition{
			Args: []string{"-e", template, "src", "dest"},
			Parameters: []*configv1.CommandLineParameterMapping{
				{
					Schema: &configv1.ParameterSchema{Name: proto.String("input")},
				},
			},
		}
		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-rsync-call")

		req := &ExecutionRequest{
			ToolName: "test-rsync",
			ToolInputs: []byte(`{"input": "sh -c 'rm -rf /'"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected")
	})

	t.Run("dangerous_env_var_prevention", func(t *testing.T) {
		cmd := "ls"
		toolDef := &v1.Tool{Name: proto.String("test-env")}
		service := &configv1.CommandLineUpstreamService{
			Command: &cmd,
		}
		// Config maps "PYTHONPATH" to input
		callDef := &configv1.CommandLineCallDefinition{
			Args: []string{"-l"},
			Parameters: []*configv1.CommandLineParameterMapping{
				{
					Schema: &configv1.ParameterSchema{Name: proto.String("PYTHONPATH")},
				},
			},
		}
		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-env-call")

		req := &ExecutionRequest{
			ToolName: "test-env",
			ToolInputs: []byte(`{"PYTHONPATH": "/tmp/malicious"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "matches a dangerous environment variable name")
	})

    t.Run("safe_env_var_allowed", func(t *testing.T) {
		cmd := "ls"
		toolDef := &v1.Tool{Name: proto.String("test-safe-env")}
		service := &configv1.CommandLineUpstreamService{
			Command: &cmd,
		}
		callDef := &configv1.CommandLineCallDefinition{
			Args: []string{"-l"},
			Parameters: []*configv1.CommandLineParameterMapping{
				{
					Schema: &configv1.ParameterSchema{Name: proto.String("MY_APP_CONFIG")},
				},
			},
		}
		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-safe-env-call")

		req := &ExecutionRequest{
			ToolName: "test-safe-env",
			ToolInputs: []byte(`{"MY_APP_CONFIG": "some_value"}`),
		}

		// It should execute (or fail at execution step, but not at validation step)
        // Since ls doesn't read stdin/args correctly here it might fail execution,
        // but we just want to ensure validation passes.
		_, err := tool.Execute(context.Background(), req)
        // If err is "matches a dangerous environment variable name", fail.
        if err != nil {
		    assert.NotContains(t, err.Error(), "matches a dangerous environment variable name")
        }
	})
}
