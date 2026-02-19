// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestPerlRCE(t *testing.T) {
	// Define a tool that uses perl
	toolDef := pb.Tool_builder{
		Name:                proto.String("perl_tool"),
		DisplayName:         proto.String("Perl Tool"),
		Description:         proto.String("Executes perl code"),
		UnderlyingMethodFqn: proto.String("perl"),
	}.Build()

	serviceConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{code}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("code"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "call1")

	// Payload: system q/echo PWNED/
	payload := `system q/echo PWNED/`

	req := &ExecutionRequest{
		ToolName:   "perl_tool",
		ToolInputs: []byte(`{"code": "` + payload + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)

	// We strictly expect an error from the security check
	if assert.Error(t, err, "Security check failed to block injection") {
		assert.Contains(t, err.Error(), "interpreter injection detected", "Error should mention interpreter injection")
		assert.Contains(t, err.Error(), "system", "Error should mention the blocked keyword")
	}
}
