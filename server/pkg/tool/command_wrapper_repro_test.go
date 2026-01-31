// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// TestLocalCommandTool_Wrapper_ShellInjection verifies that using a wrapper command like 'nice'
// still triggers shell injection checks because 'nice' is now in the isShellCommand list.
func TestLocalCommandTool_Wrapper_ShellInjection(t *testing.T) {
	t.Parallel()
	tool := v1.Tool_builder{
		Name:        proto.String("test-tool-wrapper"),
		Description: proto.String("A test tool using nice wrapper"),
	}.Build()

	// We use 'nice' which acts as a wrapper. It should be treated as a shell/dangerous command.
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("nice"),
		Local:   proto.Bool(true),
	}.Build()

	// We inject a command separator ';' to execute a second command 'echo pwned'
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"sh", "-c", "echo {{input}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "test-tool-wrapper",
		Arguments: map[string]interface{}{
			"input": "hello; echo pwned",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute the tool
	result, err := localTool.Execute(context.Background(), req)

	// Expect error because 'nice' is now in the deny list and the input contains dangerous chars
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "shell injection detected")
	assert.Contains(t, err.Error(), "dangerous character ';'")
}
