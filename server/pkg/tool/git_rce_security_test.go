// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
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
