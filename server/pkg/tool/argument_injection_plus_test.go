// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_PlusInjection_Prevention(t *testing.T) {
	t.Parallel()

	// Setup a tool definition (using vim as an example of a tool susceptible to + flags)
	toolProto := v1.Tool_builder{
		Name: proto.String("test-tool-plus"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("vim"),
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("arg")}.Build()}.Build(),
		},
		Args: []string{"{{arg}}"},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Injection attempt: Use +command syntax
	// This should be blocked by checkForArgumentInjection hardening.
	reqAttack := &ExecutionRequest{
		ToolName: "test-tool-plus",
		Arguments: map[string]interface{}{
			"arg": "+quit",
		},
	}
	reqAttack.ToolInputs, _ = json.Marshal(reqAttack.Arguments)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := localTool.Execute(ctx, reqAttack)

	// We expect an error "argument injection detected"
	if assert.Error(t, err, "Expected error blocking +flag") {
		assert.Contains(t, err.Error(), "argument injection detected", "Error should indicate argument injection")
	}
}
