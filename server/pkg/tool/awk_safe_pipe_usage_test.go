// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_AwkSafePipe_Usage(t *testing.T) {
	// Define a tool that uses 'awk'.
	tool := v1.Tool_builder{
		Name: proto.String("awk-tool"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `awk '{{script}}'`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"'{{script}}'"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: print "|"
	// This uses pipe character inside a string literal. It should be SAFE.
	// The hardened implementation should allow this while blocking unquoted pipes.
	payload := `BEGIN { print "|" }`

	req := &ExecutionRequest{
		ToolName: "awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	_, err := localTool.Execute(context.Background(), req)

	if err != nil {
		t.Fatalf("Safe awk payload %q was blocked: %v", payload, err)
	}
}
