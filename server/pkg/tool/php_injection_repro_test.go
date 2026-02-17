// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestPHPInjectionRepro(t *testing.T) {
	// Setup: Define a tool that executes php code via -r
	toolDef := v1.Tool_builder{
		Name: proto.String("php-eval"),
	}.Build()

	serviceConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("php"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-r", "'{{code}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("code"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	// Create the tool
	tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test-call-id")

	// Payload: Try to execute passthru which is currently NOT blocked
	// We use double quotes to avoid breaking out of the single-quoted argument in shell
	payload := "passthru(\"echo vulnerable\");"

	inputs := map[string]interface{}{
		"code": payload,
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "php-eval",
		ToolInputs: inputBytes,
		DryRun:     true, // DryRun runs validation but skips execution
	}

	// Execute
	_, err := tool.Execute(context.Background(), req)

	// Assert: Now we expect an error because passthru is blocked
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interpreter injection detected")
	assert.Contains(t, err.Error(), "passthru")
}
