// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCommandWrapper_Repro(t *testing.T) {
	// We verify that 'nice' is NOT considered a shell command,
	// and thus bypasses shell injection checks.

	// Check if 'nice' command exists
	path, err := exec.LookPath("nice")
	if err != nil {
		t.Skip("nice command not found")
	}

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"sh", "-c", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build(),
			}.Build(),
		},
	}.Build()

	// Create tool with "nice" command
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String(path),
	}).Build()

	// Minimal tool definition
	toolProto := v1.Tool_builder{
		Name: proto.String("nice-wrapper"),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"properties": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"script": structpb.NewStructValue(&structpb.Struct{}),
					},
				}),
			},
		},
	}.Build()

	cmdTool := tool.NewLocalCommandTool(
		toolProto,
		service,
		callDef,
		nil,
		"call-id",
	)

	// Attempt injection: simple command that prints output.
	// Payload: "echo pwned" (contains space)
	// checkUnquotedInjection blocks space.
	// If 'nice' is in the list, this should fail.
	// If 'nice' is NOT in the list, this should succeed (execute).

	inputData := map[string]interface{}{"script": "echo pwned"}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	_, err = cmdTool.Execute(context.Background(), req)

	// EXPECTATION: With the fix, this SHOULD error.
	assert.Error(t, err, "Should fail when 'nice' is in shell list (secure behavior)")
	assert.Contains(t, err.Error(), "shell injection detected")
}
