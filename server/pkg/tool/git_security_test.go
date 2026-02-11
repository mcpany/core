// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Git_C_Injection_Repro(t *testing.T) {
	// This test attempts to demonstrate that git allows command execution via -c configuration.
	// Specifically core.sshCommand.
	// Sentinel Security Update: This test now asserts that such attempts are blocked.

	toolProto := v1.Tool_builder{
		Name: proto.String("git-ssh-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
	}.Build()

	// git -c core.sshCommand={{cmd}} fetch ssh://localhost/repo
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "core.sshCommand={{cmd}}", "fetch", "ssh://localhost/repo"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("cmd"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// The payload: sh -c "echo PWNED; exit 1"
	// We want to force output to stderr/stdout
	payload := "sh -c 'echo PWNED; exit 1'"
	// We need to escape quotes for JSON
	inputs := fmt.Sprintf(`{"cmd": "%s"}`, payload)

	req := &ExecutionRequest{
		ToolName:   "git-ssh-tool",
		ToolInputs: []byte(inputs),
	}

	result, err := tool.Execute(context.Background(), req)

	// Assert that execution is BLOCKED
	if err == nil {
		t.Fatal("Expected execution to be blocked by security policy, but it succeeded.")
	}

	// Verify the error message indicates shell injection detection (due to spaces in unquoted arg)
	if !strings.Contains(err.Error(), "shell injection detected") {
		t.Errorf("Expected 'shell injection detected' error, got: %v", err)
	} else {
		t.Logf("Security check passed: %v", err)
	}

	// Double check output just in case
	var output string
	if result != nil {
		resultMap, ok := result.(map[string]interface{})
		if ok {
			if combined, ok := resultMap["combined_output"].(string); ok {
				output += combined
			}
		}
	}

	if strings.Contains(output, "PWNED") {
		t.Errorf("VULNERABILITY CONFIRMED: RCE via git -c core.sshCommand. Output: %s", output)
	}
}
