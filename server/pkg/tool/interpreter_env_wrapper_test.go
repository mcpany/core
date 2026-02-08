// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
)

func strPtrWrapper(s string) *string { return &s }

func TestEnvPythonInjection(t *testing.T) {
	// This test reproduces an RCE vulnerability where "env python" bypasses
	// interpreter-specific security checks (like f-string protection).

	// Case 1: Simple env wrapper
	// Command: env python3 -c "print(f'Hello, {{name}}')"
	t.Run("SimpleEnvWrapper", func(t *testing.T) {
		service := configv1.CommandLineUpstreamService_builder{
			Command: strPtrWrapper("env"),
		}.Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"python3", "-c", "print(f'Hello, {{name}}')"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{
						Name: strPtrWrapper("name"),
					}.Build(),
				}.Build(),
			},
		}.Build()

		toolProto := pb.Tool_builder{
			Name: strPtrWrapper("python_hello"),
		}.Build()

		tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

		payload := "{getattr(__import__(\"os\"), \"sys\"+\"tem\")(\"echo RCE_SUCCESS\")}"
		inputs := map[string]interface{}{"name": payload}
		inputsBytes, _ := json.Marshal(inputs)

		req := &ExecutionRequest{ToolName: "python_hello", ToolInputs: inputsBytes}
		_, err := tool.Execute(context.Background(), req)

		if err == nil {
			t.Fatal("VULNERABILITY CONFIRMED: Python f-string injection successful via 'env' wrapper")
		}
		assert.ErrorContains(t, err, "injection detected")
	})

	// Case 2: Env wrapper with flags
	// Command: env -i NAME=VAL python3 -c "print(f'Hello, {{name}}')"
	t.Run("EnvWithFlags", func(t *testing.T) {
		service := configv1.CommandLineUpstreamService_builder{
			Command: strPtrWrapper("env"),
		}.Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"-i", "TEST=1", "python3", "-c", "print(f'Hello, {{name}}')"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{
						Name: strPtrWrapper("name"),
					}.Build(),
				}.Build(),
			},
		}.Build()

		toolProto := pb.Tool_builder{
			Name: strPtrWrapper("python_hello_flags"),
		}.Build()

		tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

		payload := "{getattr(__import__(\"os\"), \"sys\"+\"tem\")(\"echo RCE_SUCCESS\")}"
		inputs := map[string]interface{}{"name": payload}
		inputsBytes, _ := json.Marshal(inputs)

		req := &ExecutionRequest{ToolName: "python_hello_flags", ToolInputs: inputsBytes}
		_, err := tool.Execute(context.Background(), req)

		if err == nil {
			t.Fatal("VULNERABILITY CONFIRMED: Python f-string injection successful via 'env -i' wrapper")
		}
		assert.ErrorContains(t, err, "injection detected")
	})
}
