// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestFindInjection(t *testing.T) {
	t.Run("Find_Exec_Injection", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("find_tool"),
		}).Build()
		cmd := "find"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{".", "-name", "{{filename}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("filename"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Malicious input using -exec with +
		// We avoid {}, using a command that accepts arguments (like echo or ls)
		input := "foo -exec echo pwned +"

		req := &ExecutionRequest{
			ToolName: "find_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"filename": %q}`, input)),
			Arguments: map[string]interface{}{
				"filename": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "find injection detected", "Should detect -exec injection")
	})

	t.Run("Find_Delete_Injection", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("find_delete_tool"),
		}).Build()
		cmd := "find"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{".", "-name", "{{filename}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("filename"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Malicious input using -delete
		input := "foo -delete"

		req := &ExecutionRequest{
			ToolName: "find_delete_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"filename": %q}`, input)),
			Arguments: map[string]interface{}{
				"filename": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "find injection detected", "Should detect -delete injection")
	})
}
