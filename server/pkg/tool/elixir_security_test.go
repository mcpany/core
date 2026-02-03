// Copyright 2025 Author(s) of MCP Any
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

func TestElixirSecurity(t *testing.T) {
	// Elixir Interpolation Injection in Double Quotes
	t.Run("Elixir_DoubleQuotes_Injection", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("elixir_tool"),
		}).Build()
		cmd := "elixir"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		// Template uses double quotes: IO.puts("{{msg}}")
		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{"-e", "IO.puts(\"{{msg}}\")"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("msg"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Malicious input using Elixir interpolation
		input := "#{System.halt(1)}"

		req := &ExecutionRequest{
			ToolName: "elixir_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"msg": "%s"}`, input)),
			Arguments: map[string]interface{}{
				"msg": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		// Assert that we detect the attempt
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "elixir/ruby interpolation injection detected", "Should detect elixir interpolation")
	})
}
