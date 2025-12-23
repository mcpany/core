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
	"github.com/stretchr/testify/require"
)

func TestCommandTool_ShellInjectionPrevention(t *testing.T) {
	// Setup
	// We want to run: sh -c "echo {{msg}}"
	// Payload: "hello; echo pwned" - should be blocked

	toolDef := &v1.Tool{
		Name: strPtr("echo_tool"),
	}

	service := &configv1.CommandLineUpstreamService{
		Command: strPtr("sh"),
		// No container environment means local execution
	}

	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-c", "echo {{msg}}"},
	}

	cmdTool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call-id")

	// Inputs
	inputs := map[string]interface{}{
		"msg": "hello; echo pwned",
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "echo_tool",
		ToolInputs: json.RawMessage(inputBytes),
	}

	// Execute
	_, err := cmdTool.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "shell injection attempt detected")
}
