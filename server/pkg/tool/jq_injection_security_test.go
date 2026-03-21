// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestJQEnvLeakage(t *testing.T) {
	// Setup service config
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("jq"),
	}.Build()

	// Setup call definition
	// jq -n '{{input}}'
    // We rely on the fact that if we pass unquoted argument, it is not wrapped in quotes by the code,
    // but checks are applied assuming it is unquoted.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-n", "{{input}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	inputProperties, _ := structpb.NewStruct(map[string]interface{}{
		"input": map[string]interface{}{"type": "string"},
	})

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type":       structpb.NewStringValue("object"),
			"properties": structpb.NewStructValue(inputProperties),
		},
	}

	toolProto := pb.Tool_builder{
		Name:        proto.String("jq_leak"),
		InputSchema: inputSchema,
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "test")

	// Payload: env
	payload := "env"

	inputs := map[string]interface{}{
		"input": payload,
	}
	inputsBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "jq_leak",
		ToolInputs: inputsBytes,
	}

	_, err := tool.Execute(context.Background(), req)

	if err == nil {
		t.Log("VULNERABILITY: Validation passed for jq env leakage")
		t.Fail()
	} else {
		if assert.NotContains(t, err.Error(), "executable file not found") {
			t.Logf("Got expected error: %v", err)
			assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
		} else {
			t.Fatal("VULNERABILITY: Validation passed (attempted execution)")
		}
	}
}
