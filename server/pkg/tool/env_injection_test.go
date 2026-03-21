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

func TestEnvInjection_Repro(t *testing.T) {
	cmd := "env"
	tool := createEnvCommandTool(cmd)
	req := &ExecutionRequest{
		ToolName: "test",
		ToolInputs: []byte(`{"input": "LD_PRELOAD=/tmp/evil.so"}`),
	}

	_, err := tool.Execute(context.Background(), req)
	// This currently passes (nil error) but should fail
	if err == nil {
		t.Log("Vulnerability confirmed: `=` was allowed in input for `env` command")
	} else {
		t.Logf("Blocked with error: %v", err)
	}

	// We want to assert that it IS an error
	assert.Error(t, err, "Should detect variable injection using '='")
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected", "Error message should mention shell injection")
	}
}

func createEnvCommandTool(command string) Tool {
	toolDef := v1.Tool_builder{Name: proto.String("test-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &command,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{input}}", "sh", "-c", "echo hello"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()
	// Policies nil, callID "test-call"
	return NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
}
