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
	"google.golang.org/protobuf/proto"
)

func TestInterpreter_Usability_Regression(t *testing.T) {
	// This test ensures that tools like python allow structured data (JSON, URLs)
	// inside quotes, which was previously blocked by strict mode.

	toolProto := v1.Tool_builder{
		Name: proto.String("python-script"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("data")}.Build()}.Build(),
		},
		Args: []string{"script.py", "'{{data}}'"},
	}.Build()

	cmdTool := tool.NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Case 1: JSON Data (contains { } " :)
	// Should be ALLOWED
	jsonData := `{"key": "value", "list": [1, 2]}`
	inputs, _ := json.Marshal(map[string]interface{}{"data": jsonData})

	// We expect Execute to proceed past validation.

	_, err := cmdTool.Execute(context.Background(), &tool.ExecutionRequest{ToolInputs: inputs})

	if err != nil {
		assert.NotContains(t, err.Error(), "shell injection detected", "JSON data inside quotes should be allowed for python")
	}

	// Case 2: URL with query params (contains ? & =)
	urlData := "https://example.com/search?q=foo&lang=en"
	inputsURL, _ := json.Marshal(map[string]interface{}{"data": urlData})

	_, err = cmdTool.Execute(context.Background(), &tool.ExecutionRequest{ToolInputs: inputsURL})
	if err != nil {
		assert.NotContains(t, err.Error(), "shell injection detected", "URL data inside quotes should be allowed for python")
	}
}

func TestShell_StrictValidation(t *testing.T) {
	// Verify that sh still blocks dangerous chars
	toolProto := v1.Tool_builder{Name: proto.String("sh-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("cmd")}.Build()}.Build(),
		},
		Args: []string{"-c", "'{{cmd}}'"},
	}.Build()

	cmdTool := tool.NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Dangerous input
	danger := "echo hello; rm -rf /"
	inputs, _ := json.Marshal(map[string]interface{}{"cmd": danger})

	_, err := cmdTool.Execute(context.Background(), &tool.ExecutionRequest{ToolInputs: inputs})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shell injection detected", "Shell should block dangerous chars even in quotes")
}
