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

func TestPerlExecInjection(t *testing.T) {
	t.Run("Perl_Exec_NoParen_Injection", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("perl_tool"),
		}).Build()
		cmd := "perl"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		// Template: perl -e '{{code}}'
		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{"-e", "'{{code}}'"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("code"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Malicious input: exec "echo injected"
		// This should be blocked, but currently passes because it doesn't use parentheses
		input := `exec "echo injected"`

		req := &ExecutionRequest{
			ToolName: "perl_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"code": %q}`, input)),
			Arguments: map[string]interface{}{
				"code": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		// We expect this to fail with "injection detected"
		if err == nil {
			t.Fatal("VULNERABILITY CONFIRMED: Perl exec injection allowed")
		} else {
			assert.Contains(t, err.Error(), "injection detected")
		}
	})

	t.Run("Perl_Exec_Bareword_Injection", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("perl_tool_bareword"),
		}).Build()
		cmd := "perl"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		// Template: perl -e '{{code}}'
		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{"-e", "'{{code}}'"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("code"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Malicious input: exec echo
		input := `exec echo`

		req := &ExecutionRequest{
			ToolName: "perl_tool_bareword",
			ToolInputs: []byte(fmt.Sprintf(`{"code": %q}`, input)),
			Arguments: map[string]interface{}{
				"code": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		// We expect this to fail with "injection detected"
		if err == nil {
			t.Fatal("VULNERABILITY CONFIRMED: Perl bareword exec injection allowed")
		} else {
			assert.Contains(t, err.Error(), "interpreter injection detected")
		}
	})
}
