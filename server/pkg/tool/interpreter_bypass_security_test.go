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

func TestPHPPassThruInjection(t *testing.T) {
	// Setup service config
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("php"),
	}.Build()

	// Setup call definition
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-r", "echo '{{input}}';"},
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
		Name:        proto.String("php_passthru"),
		InputSchema: inputSchema,
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "test")

	payload := "'; passthru('echo RCE'); //"

	req := &ExecutionRequest{
		ToolName:   "php_passthru",
		ToolInputs: []byte(`{"input": "` + payload + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)

	if err == nil {
		t.Log("VULNERABILITY: Validation passed for PHP passthru injection")
		t.Fail()
	} else {
		t.Logf("Got expected error: %v", err)
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
	}
}

func TestRubySyscallInjection(t *testing.T) {
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "puts '{{input}}'"},
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
		Name:        proto.String("ruby_syscall"),
		InputSchema: inputSchema,
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "test")

	payload := "'; syscall(20); #"

	req := &ExecutionRequest{
		ToolName:   "ruby_syscall",
		ToolInputs: []byte(`{"input": "` + payload + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)

	if err == nil {
		t.Log("VULNERABILITY: Validation passed for Ruby syscall injection")
		t.Fail()
	} else {
		t.Logf("Got expected error: %v", err)
		assert.Contains(t, err.Error(), "injection detected", "Expected injection detection error")
	}
}
