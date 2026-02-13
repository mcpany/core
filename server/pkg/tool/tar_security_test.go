// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
    "github.com/stretchr/testify/assert"
)

func TestLocalCommandTool_Tar_Checkpoint_Security(t *testing.T) {
    // This test verifies that tar command execution via --checkpoint-action is protected
    // against shell injection.

	toolProto := v1.Tool_builder{
		Name: proto.String("tar-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("tar"),
		Local:   proto.Bool(true),
	}.Build()

    // tar cf archive.tar --checkpoint-action=exec={{cmd}} file
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"cf", "archive.tar", "--checkpoint-action=exec={{cmd}}", "file"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("cmd"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

    // Payload attempting to inject a shell command via spaces
	payload := "sh -c 'echo VULN'"
	inputs := fmt.Sprintf(`{"cmd": "%s"}`, payload)

	req := &ExecutionRequest{
		ToolName:   "tar-tool",
		ToolInputs: []byte(inputs),
	}

	result, err := tool.Execute(context.Background(), req)

    assert.Error(t, err)
    assert.Nil(t, result)
    if err != nil {
        assert.Contains(t, err.Error(), "shell injection detected")
        assert.Contains(t, err.Error(), "contains dangerous character ' '")
    }
}
