// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestCommandInjection_AwkRepro demonstrates that strict validation is needed even inside quotes
// when the command is an interpreter.
func TestCommandInjection_AwkRepro(t *testing.T) {
	// Setup a LocalCommandTool that uses bash to invoke awk
	// usage: bash -c "awk '{{script}}' /dev/null"
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-c", "awk '{{script}}' /dev/null"},
		Parameters: []*configv1.CommandLineParameterMapping{
			{Schema: &configv1.ParameterSchema{Name: proto.String("script")}},
		},
	}

	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("bash"),
		Local:   proto.Bool(true),
	}).Build()

	properties := map[string]*structpb.Value{
		"script": structpb.NewStructValue(&structpb.Struct{}),
	}
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: properties,
			}),
		},
	}

	// Create LocalCommandTool directly to simulate local execution
	toolProto := (&pb.Tool_builder{
		Name:        proto.String("awk-tool"),
		InputSchema: inputSchema,
	}).Build()

	cmdTool := tool.NewLocalCommandTool(
		toolProto,
		service,
		callDef,
		nil,
		"call-id",
	)

	// Input that executes code: BEGIN { print "pwned" }
	// This input contains { } " which are dangerous.
	// Previously, inside single quotes, this was allowed.
	// Now, it should be blocked.
	inputData := map[string]interface{}{"script": "BEGIN { print \"pwned\" }"}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)

	req := &tool.ExecutionRequest{ToolInputs: inputs}

	result, err := cmdTool.Execute(context.Background(), req)

	// After the fix, this should FAIL with a security error.
	assert.Error(t, err, "Expected security error")
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected", "Error message should mention shell injection")
		assert.Contains(t, err.Error(), "enforcing strict mode", "Error message should mention strict mode")
	}

	if err == nil {
		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		stdout := resultMap["stdout"].(string)
		t.Logf("Output: %s", stdout)
		assert.Fail(t, "Vulnerability confirmed: Code was executed (should have been blocked)")
	}
}
