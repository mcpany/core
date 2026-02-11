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
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_GitConfigInjection_Security(t *testing.T) {
	// This test attempts to demonstrate that git allows command execution via -c configuration.
	// Specifically core.sshCommand. We expect this to be BLOCKED by our security enhancements.

	toolProto := v1.Tool_builder{
		Name: proto.String("git-ssh-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
	}.Build()

	// git -c core.sshCommand={{cmd}} clone ssh://localhost/repo
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "core.sshCommand={{cmd}}", "clone", "ssh://localhost/repo"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("cmd"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// The payload: sh -c "echo PWNED"
	// This contains spaces, which should be blocked by checkUnquotedInjection because
	// git is now marked as isConfigInjectionSensitive.
	payload := "sh -c 'echo PWNED; exit 1'"
	// We need to escape quotes for JSON
	inputs := fmt.Sprintf(`{"cmd": "%s"}`, payload)

	req := &ExecutionRequest{
		ToolName:   "git-ssh-tool",
		ToolInputs: []byte(inputs),
	}

	result, err := tool.Execute(context.Background(), req)

	// Sentinel Security Update: We expect this to fail now because 'git' is treated as sensitive
	// and strict checking blocks spaces in unquoted arguments (config substitution).
	if err == nil {
		t.Errorf("Expected error 'shell injection detected', but got nil. Result: %v", result)
	} else {
		if !strings.Contains(err.Error(), "shell injection detected") {
			t.Errorf("Expected error 'shell injection detected', but got: %v", err)
		} else {
			// Success
		}
	}

	var output string
	if result != nil {
		resultMap, ok := result.(map[string]interface{})
		if ok {
			if combined, ok := resultMap["combined_output"].(string); ok {
				output += combined
			}
		}
	}

	if strings.Contains(output, "PWNED") || strings.Contains(output, "PWNE") {
		t.Fatalf("VULNERABILITY STILL EXISTS: RCE via git -c core.sshCommand. Output: %s", output)
	}
}

func TestLocalCommandTool_Git_Usability(t *testing.T) {
	// Verify that we can still use spaces in arguments if passed via "args" array (Dynamic Args),
	// which skips the strict "Config Injection" check.

	toolProto := v1.Tool_builder{
		Name: proto.String("git-status-tool"),
		InputSchema: (&structpb.Struct{
			Fields: map[string]*structpb.Value{
				"properties": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"args": structpb.NewStructValue(&structpb.Struct{}),
					},
				}),
			},
		}),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
	}.Build()

	// Tool with NO fixed args, relies on dynamic args
	callDef := configv1.CommandLineCallDefinition_builder{}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// git status "my folder"
	inputs := `{"args": ["status", "my folder"]}`

	req := &ExecutionRequest{
		ToolName:   "git-status-tool",
		ToolInputs: []byte(inputs),
	}

	result, err := tool.Execute(context.Background(), req)

	// We expect NO "shell injection detected" error.
	if err != nil {
		if strings.Contains(err.Error(), "shell injection detected") {
			t.Errorf("Usability Regression: 'args' array blocked spaces: %v", err)
		} else {
			// Other errors are fine (e.g. exit status 128)
			t.Logf("Execution error (expected): %v", err)
		}
	} else {
		t.Logf("Execution success: %v", result)
	}
}
