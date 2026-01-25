// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_WindowsInjection_Prevention(t *testing.T) {
	t.Parallel()

	tool := &v1.Tool{Name: proto.String("test-tool-cmd")}
	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("cmd.exe"),
		Local:   proto.Bool(true),
		ContainerEnvironment: &configv1.ContainerEnvironment{
			Image: proto.String("windows-image"), // Simulate Docker to bypass absolute path check
		},
	}
	callDef := &configv1.CommandLineCallDefinition{
		Parameters: []*configv1.CommandLineParameterMapping{
			{Schema: &configv1.ParameterSchema{Name: proto.String("arg")}},
		},
		Args: []string{"{{arg}}"},
	}

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Injection attempt: Use /c to execute a command
	reqAttack := &ExecutionRequest{
		ToolName: "test-tool-cmd",
		Arguments: map[string]interface{}{
			"arg": "/c calc",
		},
	}
	reqAttack.ToolInputs, _ = json.Marshal(reqAttack.Arguments)

	_, err := localTool.Execute(context.Background(), reqAttack)

	// The security check should catch this BEFORE trying to execute.
	// If it tries to execute, it will likely fail with "executable file not found" (on Linux)
	// or actually run calc (on Windows). Both are failures of the security check.
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			t.Fatal("Security check bypassed! Attempted to execute cmd.exe with injected argument.")
		}
		// If it's a security error, we expect something like "argument injection" or "shell injection"
		assert.Contains(t, err.Error(), "injection", "Expected injection error")
	} else {
		t.Fatal("Security check bypassed! Command executed successfully (unexpectedly).")
	}
}
