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

func TestInterpreterSecurity_SubprocessBypass(t *testing.T) {
	// Python Subprocess Injection Bypass
	// This tests if `subprocess.call` can be used to bypass the keyword blocking
	// because `subprocess` is followed by `.` which is not checked as a delimiter.
	t.Run("Python_Subprocess_Bypass", func(t *testing.T) {
		toolDef := (&pb.Tool_builder{
			Name: proto.String("python_tool"),
		}).Build()
		cmd := "python3"
		serviceConfig := (&configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}).Build()

		// Use quoted argument to allow ; and (
		callDef := (&configv1.CommandLineCallDefinition_builder{
			Args: []string{"-c", "'{{code}}'"},
			Parameters: []*configv1.CommandLineParameterMapping{
				(&configv1.CommandLineParameterMapping_builder{
					Schema: (&configv1.ParameterSchema_builder{
						Name: proto.String("code"),
					}).Build(),
				}).Build(),
			},
		}).Build()

		tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test_call")

		// Malicious input using subprocess.call
		// import subprocess; subprocess.call(["echo", "pwned"])
		// NOTE: We use double quotes inside to avoid breaking the single quoted argument in the template
		input := "import subprocess; subprocess.call([\"echo\", \"pwned\"])"

		req := &ExecutionRequest{
			ToolName: "python_tool",
			// ToolInputs is JSON. We need to be careful with escaping.
			// `{"code": "import ... \"echo\" ..."}`
			ToolInputs: []byte(fmt.Sprintf(`{"code": %q}`, input)),
			Arguments: map[string]interface{}{
				"code": input,
			},
		}

		_, err := tool.Execute(context.Background(), req)

		// Assert that it IS blocked
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "interpreter injection detected", "Should detect subprocess call injection")
	})
}
