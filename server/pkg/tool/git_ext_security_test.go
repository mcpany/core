// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_Git_Ext_RCE_Repro(t *testing.T) {
	// This test attempts to demonstrate that git command execution allows injecting
	// dangerous protocols like ext:: which can lead to RCE.

    // Setup temporary file for proof
    tmpFile := "/tmp/pwned_git_ext"
    os.Remove(tmpFile)
    defer os.Remove(tmpFile)

	tool := configv1.ToolDefinition_builder{
		Name: proto.String("git-clone"),
	}.Build()

    toolProto := v1.Tool_builder{
        Name: proto.String("git-clone"),
    }.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
        Tools:   []*configv1.ToolDefinition{tool},
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"clone", "{{url}}", "/tmp/repo"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("url"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// The payload: ext::sh -c 'touch /tmp/pwned_git_ext'
    // We use quotes to ensure arguments are passed correctly to sh -c
	payload := "ext::sh -c 'touch /tmp/pwned_git_ext'"

	req := &ExecutionRequest{
		ToolName: "git-clone",
		Arguments: map[string]interface{}{
			"url": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

	_, err := localTool.Execute(ctx, req)

    // Check if file was created (should not be)
    if _, statErr := os.Stat(tmpFile); statErr == nil {
		t.Errorf("VULNERABILITY CONFIRMED: RCE via git ext:: protocol. File %s created.", tmpFile)
	}

    // We expect an error from the tool execution blocking the ext:: protocol
    if err == nil {
        t.Errorf("Expected error blocking ext:: protocol, but got nil (SECURITY REGRESSION)")
    } else if strings.Contains(err.Error(), "ext:: scheme detected") {
        t.Logf("SUCCESS: Blocked ext:: protocol with error: %v", err)
    } else {
        t.Errorf("Got error but not specific blocking message: %v", err)
    }
}
