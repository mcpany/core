// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_Leakage(t *testing.T) {
	// This test verifies that sensitive inputs are redacted from args and output.
	// "api_key" is a known sensitive key in util.IsSensitiveKey

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"api_key": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}

	tool := v1.Tool_builder{
		Name:        proto.String("leak-test-tool"),
		Description: proto.String("A test tool"),
		InputSchema: inputSchema,
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("api_key"),
				}.Build(),
			}.Build(),
		},
		Args: []string{"{{api_key}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	secretVal := "SuperSecretValue123"
	// JSON input
	req := &ExecutionRequest{
		ToolName:   "leak-test-tool",
		ToolInputs: []byte(`{"api_key": "` + secretVal + `"}`),
	}

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// Check args
	args, ok := resultMap["args"].([]string)
	assert.True(t, ok)

	// Check stdout
	stdout, ok := resultMap["stdout"].(string)
	assert.True(t, ok)

	// Assertions - SHOULD PASS after fix
	foundInArgs := false
	for _, arg := range args {
		if arg == secretVal {
			foundInArgs = true
		}
	}
	assert.False(t, foundInArgs, "Args should not contain secret")

	assert.NotContains(t, stdout, secretVal, "Stdout should not contain secret")
	assert.Contains(t, stdout, "[REDACTED]", "Stdout should contain redaction placeholder")
}
