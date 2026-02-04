// Copyright 2025 Author(s) of MCP Any
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

func TestLocalCommandTool_Execute_GitCommitMessageWithSpaces(t *testing.T) {
	t.Parallel()
	tool := v1.Tool_builder{
		Name: proto.String("git-commit"),
	}.Build()
	// git is removed from isShellCommand list
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"commit", "-m", "{{message}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("message")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "git-commit",
		Arguments: map[string]interface{}{
			"message": "Initial commit with spaces",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// This should now succeed because git is no longer considered a shell command
	_, err := localTool.Execute(context.Background(), req)

	assert.NoError(t, err, "Expected no error for git command with spaces in arguments")
}
