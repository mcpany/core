// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Git_RCE_Repro(t *testing.T) {
	// This test attempts to demonstrate that git command execution allows injecting
	// dangerous environment variables like GIT_SSH_COMMAND which can lead to RCE.

	tool := v1.Tool_builder{
		Name: proto.String("git-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
	}.Build()

	// Use git clone with SSH URL to trigger SSH command execution
	// Use localhost to fail fast if SSH is attempted
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"clone", "ssh://localhost:12345/repo.git"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("GIT_SSH_COMMAND"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// The payload: sh -c echo PWNED
	payload := "sh -c echo PWNED"

	req := &ExecutionRequest{
		ToolName: "git-tool",
		Arguments: map[string]interface{}{
			"GIT_SSH_COMMAND": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)

	var output string
	if err != nil {
		output = err.Error()
	}

	if result != nil {
		resultMap, ok := result.(map[string]interface{})
		if ok {
			if combined, ok := resultMap["combined_output"].(string); ok {
				output += combined
			}
		}
	}

	if strings.Contains(output, "PWNED") {
		t.Errorf("VULNERABILITY CONFIRMED: RCE via GIT_SSH_COMMAND. Output: %s", output)
	} else {
		// Log success but don't clutter output
		// t.Logf("Vulnerability not triggered (Safe). Output: %s", output)
	}
}

func TestLocalCommandTool_Git_Pager_RCE(t *testing.T) {
	// This test demonstrates RCE via PAGER environment variable in git.
	// PAGER is executed by git to display output if pagination is forced.

	// Mock IsAllowedPath to allow temp dir
	originalIsAllowedPath := validation.IsAllowedPath
	defer func() { validation.IsAllowedPath = originalIsAllowedPath }()
	validation.IsAllowedPath = func(path string) error {
		return nil
	}

	// Setup temp git repo
	dir := t.TempDir()
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	tool := v1.Tool_builder{
		Name: proto.String("git-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command:          proto.String("git"),
		Local:            proto.Bool(true),
		WorkingDirectory: proto.String(dir),
	}.Build()

	// 'git var GIT_PAGER' prints the pager that will be used.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"var", "GIT_PAGER"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("PAGER"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	policies := []*configv1.CallPolicy{}
	localTool := NewLocalCommandTool(tool, service, callDef, policies, "call-id")

	// Payload: execute 'echo PWNED' via sh
	payload := "sh -c 'echo PWNED'"

	req := &ExecutionRequest{
		ToolName: "git-tool",
		ToolInputs: []byte(`{"PAGER": "sh -c 'echo PWNED'"}`),
		Arguments: map[string]interface{}{
			"PAGER": payload,
		},
	}

	result, err := localTool.Execute(context.Background(), req)

	var output string
	if err != nil {
		output = err.Error()
	}

	if result != nil {
		resultMap, ok := result.(map[string]interface{})
		if ok {
			if combined, ok := resultMap["combined_output"].(string); ok {
				output += combined
			}
		}
	}

	t.Logf("Output: %s", output)

	if strings.Contains(output, "PWNED") {
		t.Errorf("VULNERABILITY CONFIRMED: RCE via PAGER. Output: %s", output)
	} else {
		t.Logf("SUCCESS: PAGER injection blocked. Output: %s", output)
	}
}
