// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestJqInjection(t *testing.T) {
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("jq"),
	}.Build()

	// Simulate jq -n "{{input}}"
	// input is inside double quotes in the jq filter
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-n", "\"{{input}}\""},
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
		Name:        proto.String("jq_injection"),
		InputSchema: inputSchema,
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "test")

	// Payload: \(env.USER)
	// This uses jq string interpolation to read environment variable
	// We need to double escape backslash for JSON string: "\\(env.USER)" -> Go: "\\\\(env.USER)"
	payload := "\\\\(env.USER)"

	req := &ExecutionRequest{
		ToolName:   "jq_injection",
		ToolInputs: []byte(`{"input": "` + payload + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)

	if err == nil {
		t.Log("VULNERABILITY: Validation passed for jq interpolation injection")
		t.Fail()
	} else {
		t.Logf("Got expected error: %v", err)
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
	}
}
