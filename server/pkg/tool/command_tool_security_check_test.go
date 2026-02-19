// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestCommandTool_ShellInjection_Prevention verifies that CommandTool
// enforces shell injection checks on the 'args' parameter when executing locally.
func TestCommandTool_ShellInjection_Prevention(t *testing.T) {
	t.Parallel()

	t.Run("awk system injection", func(t *testing.T) {
		// Setup a CommandTool configured to use "awk"
		// CommandTool executes locally if no container env is set.

		service := (&configv1.CommandLineUpstreamService_builder{
			Command: proto.String("awk"),
		}).Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
				}.Build(),
			},
		}.Build()

		inputSchema := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"properties": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"args": structpb.NewStructValue(&structpb.Struct{}),
					},
				}),
			},
		}

		toolProto := v1.Tool_builder{
			Name:        proto.String("awk-rce-test"),
			InputSchema: inputSchema,
		}.Build()

		cmdTool := tool.NewCommandTool(
			toolProto,
			service,
			callDef,
			nil,
			"call-id",
		)

		// Payload: awk 'BEGIN { print "RCE_SUCCESS" }'
		// args: ["BEGIN { system(\"echo RCE_SUCCESS\") }"]
		// This contains '(', which is dangerous in unquoted context.

		inputData := map[string]interface{}{
			"args": []string{"BEGIN { system(\"echo RCE_SUCCESS\") }"},
		}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)

		req := &tool.ExecutionRequest{ToolInputs: inputs}

		// Execute
		_, err = cmdTool.Execute(context.Background(), req)

		// Expect error due to shell injection detection
		require.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected")
	})

	t.Run("env space injection", func(t *testing.T) {
		// Verify that space is blocked for shell commands (env is treated as shell)
		service := (&configv1.CommandLineUpstreamService_builder{
			Command: proto.String("env"),
		}).Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
				}.Build(),
			},
		}.Build()

		inputSchema := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"properties": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"args": structpb.NewStructValue(&structpb.Struct{}),
					},
				}),
			},
		}

		cmdTool := tool.NewCommandTool(
			v1.Tool_builder{
				Name:        proto.String("env-test"),
				InputSchema: inputSchema,
			}.Build(),
			service,
			callDef,
			nil,
			"call-id",
		)

		inputData := map[string]interface{}{"args": []string{"echo", "hello world"}}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)

		req := &tool.ExecutionRequest{ToolInputs: inputs}
		_, err = cmdTool.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected")
	})
}
