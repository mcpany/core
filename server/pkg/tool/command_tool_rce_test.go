// Copyright 2026 Author(s) of MCP Any
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
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCommandTool_RCE_Vulnerability(t *testing.T) {
	// Setup a CommandTool configured to use "python3" but WITHOUT local: true
	// This forces it to use CommandTool instead of LocalCommandTool.

	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		// Local is implicitly false
	}).Build()

	// Define the call definition allowing "args" and having "-c" hardcoded
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
			}.Build(),
		},
	}.Build()

	// Allow "args" in input schema
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"args": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}

	toolProto := v1.Tool_builder{
		Name:        proto.String("rce-test"),
		InputSchema: inputSchema,
	}.Build()

	cmdTool := tool.NewCommandTool(
		toolProto,
		service,
		callDef,
		nil,
		"call-id",
	)

	// Payload: print("INJECTED")
	// If CommandTool was using LocalCommandTool logic, it would block " because checkUnquotedInjection blocks it.

	inputData := map[string]interface{}{
		"args": []string{"print(\"INJECTED\")"},
	}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)

	req := &tool.ExecutionRequest{ToolInputs: inputs}

	// Execute
	_, err = cmdTool.Execute(context.Background(), req)

	// Expect error because we blocked the injection
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "shell injection detected")
		t.Log("Vulnerability Mitigated: Code injection blocked as expected")
	} else {
		t.Fatal("Vulnerability Still Present: Code injection succeeded without error")
	}
}
