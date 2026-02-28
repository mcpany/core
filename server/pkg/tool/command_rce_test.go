// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"
	"encoding/json"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_RCE_Verification(t *testing.T) {
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"), // echo is typically safe
		Tools: []*configv1.ToolDefinition{
			configv1.ToolDefinition_builder{
				Name:   proto.String("test-rce"),
				CallId: proto.String("call-id"),
			}.Build(),
		},
		Calls: map[string]*configv1.CommandLineCallDefinition{
			"call-id": configv1.CommandLineCallDefinition_builder{
				Id: proto.String("call-id"),
				Args: []string{"{{text}}"},
				Parameters: []*configv1.CommandLineParameterMapping{
					configv1.CommandLineParameterMapping_builder{
						Schema: configv1.ParameterSchema_builder{
							Name:       proto.String("text"),
							IsRequired: proto.Bool(true),
						}.Build(),
					}.Build(),
				},
			}.Build(),
		},
	}.Build()

	toolDef := v1.Tool_builder{
		Name: proto.String("test-rce"),
	}.Build()

	localTool := NewLocalCommandTool(toolDef, service, service.GetCalls()["call-id"], nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "test-rce",
		Arguments: map[string]interface{}{
			"text": "hello $(whoami)",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)
	if err != nil {
		t.Logf("Execution blocked (error): %v", err)
	} else {
		t.Log("Execution succeeded.")
	}
}
