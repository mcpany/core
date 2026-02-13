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

func TestTarInjection_Vulnerability(t *testing.T) {
	// This test demonstrates that "tar" IS treated as a dangerous command,
	// preventing argument injection that can lead to RCE via --checkpoint-action.

	// Verify tar IS considered a shell command (or rather, an interpreter, which implies shell checks)
	assert.True(t, isShellCommand("tar"), "tar MUST be considered a shell command/interpreter now")

	toolDef := v1.Tool_builder{
		Name: proto.String("tar-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("tar"),
		Local:   proto.Bool(true),
	}.Build()

	// Configuration where the user input is substituted into a flag
	callDefVulnerable := configv1.CommandLineCallDefinition_builder{
		Args: []string{"cf", "archive.tar", "--checkpoint-action={{action}}", "file.txt"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("action")}.Build(),
			}.Build(),
		},
	}.Build()

	localToolVulnerable := NewLocalCommandTool(toolDef, service, callDefVulnerable, nil, "call-id")

	// 1. Malicious Payload
	req := &ExecutionRequest{
		ToolName: "tar-tool",
		Arguments: map[string]interface{}{
			"action": "exec=sh", // Malicious payload
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	_, err := localToolVulnerable.Execute(context.Background(), req)

	// Now we expect a security error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tar injection detected")

	// 2. Safe Payload
	reqSafe := &ExecutionRequest{
		ToolName: "tar-tool",
		Arguments: map[string]interface{}{
			"action": "dot", // Safe payload
		},
	}
	reqSafe.ToolInputs, _ = json.Marshal(reqSafe.Arguments)

	// Execute Safe
	_, errSafe := localToolVulnerable.Execute(context.Background(), reqSafe)

	// Should not be a security error.
	// It might fail because tar is not installed or invalid usage, but NOT injection error.
	if errSafe != nil {
		assert.NotContains(t, errSafe.Error(), "injection detected")
		assert.NotContains(t, errSafe.Error(), "blocked")
	}
}
