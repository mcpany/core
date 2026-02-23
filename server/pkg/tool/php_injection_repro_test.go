// Copyright 2026 Author(s) of MCP Any
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

func TestPHPInjection_Repro(t *testing.T) {
	t.Run("php_passthru_injection", func(t *testing.T) {
		cmd := "php"
		template := "'{{input}}'"

		toolDef := v1.Tool_builder{Name: proto.String("test-tool")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"-r", template},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
				}.Build(),
			},
		}.Build()
		tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

		req := &ExecutionRequest{
			ToolName: "test",
		}

		// Correctly escape the JSON input string for the test setup
		// input string in Go: passthru("echo RCE_SUCCESS");
		// JSON string: "passthru(\"echo RCE_SUCCESS\");"
		jsonInput := `{"input": "passthru(\"echo RCE_SUCCESS\");"}`
		req.ToolInputs = []byte(jsonInput)

		_, err := tool.Execute(context.Background(), req)

		// This assertion should FAIL if the vulnerability exists (err is nil)
		assert.Error(t, err, "Expected security error for dangerous PHP function passthru()")
		if err != nil {
			assert.Contains(t, err.Error(), "interpreter injection detected")
		}
	})
}
