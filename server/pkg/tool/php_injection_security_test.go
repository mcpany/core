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

func TestPHPInjectionSecurity(t *testing.T) {
	t.Run("PHP_Passthru_Injection", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("php_tool"),
		}).Build()
		cmd := "php"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		// Sentinel Security Update: Use double quotes to allow single quotes inside
		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{"-r", "\"{{code}}\""},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("code"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		policies := []*configv1.CallPolicy{}
		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, policies, "test_call")

		// Malicious input using passthru
		input := "passthru('echo vulnerable');"

		req := &ExecutionRequest{
			ToolName: "php_tool",
			ToolInputs: []byte(fmt.Sprintf(`{"code": %q}`, input)),
			Arguments: map[string]interface{}{
				"code": input,
			},
			DryRun: true,
		}

		_, err := tool.Execute(context.Background(), req)

		// This test asserts that the injection IS DETECTED (returns an error).
		// Before the fix, this assertion is expected to FAIL (err will be nil).
		assert.Error(t, err, "Should detect php passthru injection")
		if err != nil {
			assert.Contains(t, err.Error(), "interpreter injection detected", "Should contain specific error message")
		}
	})
}
