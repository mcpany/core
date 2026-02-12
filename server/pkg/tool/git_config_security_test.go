// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Git_C_Injection_Repro(t *testing.T) {
	// This test attempts to demonstrate that git allows command execution via -c configuration.
	// Specifically core.sshCommand.

	toolProto := v1.Tool_builder{
		Name: proto.String("git-ssh-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
	}.Build()

	// git -c core.sshCommand={{ssh_cmd}} clone ssh://localhost/foo
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "core.sshCommand={{ssh_cmd}}", "clone", "ssh://localhost/foo"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("ssh_cmd"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	t.Run("Whitespace Injection", func(t *testing.T) {
		// The payload: sh -c "touch pwned"
		payload := "sh -c 'touch pwned'"
		inputs := fmt.Sprintf(`{"ssh_cmd": "%s"}`, payload)

		req := &ExecutionRequest{
			ToolName:   "git-ssh-tool",
			ToolInputs: []byte(inputs),
		}

		// Ensure clean state
		os.Remove("pwned")
		defer os.Remove("pwned")

		result, err := tool.Execute(context.Background(), req)

		// We EXPECT an error now due to security fix
		if err == nil {
			t.Error("Expected security error, but got nil")
		} else {
			errStr := err.Error()
			if strings.Contains(errStr, "with whitespace") {
				t.Logf("Security check passed: %v", err)
			} else {
				t.Errorf("Expected whitespace error, got: %v", err)
			}
		}

		if result != nil {
             t.Logf("Result: %+v", result)
        }
	})

	t.Run("No-Space Injection", func(t *testing.T) {
		// The payload: sh
		payload := "sh"
		inputs := fmt.Sprintf(`{"ssh_cmd": "%s"}`, payload)

		req := &ExecutionRequest{
			ToolName:   "git-ssh-tool",
			ToolInputs: []byte(inputs),
		}

		result, err := tool.Execute(context.Background(), req)

		// We EXPECT an error now due to security fix
		if err == nil {
			t.Error("Expected security error, but got nil")
		} else {
			errStr := err.Error()
			if strings.Contains(errStr, "suspicious git config value") {
				t.Logf("Security check passed: %v", err)
			} else {
				t.Errorf("Expected config value error, got: %v", err)
			}
		}

		if result != nil {
             t.Logf("Result: %+v", result)
        }
	})
}
