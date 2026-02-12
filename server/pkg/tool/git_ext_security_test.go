// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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
    tmpFile := filepath.Join(t.TempDir(), "pwned_git_ext")

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

	// The payload: ext::sh -c "touch /tmp/pwned_git_ext"
    // We use a harmless command but verify execution via file creation
	payload := "ext::sh -c touch /tmp/pwned_git_ext"

	req := &ExecutionRequest{
		ToolName: "git-clone",
		Arguments: map[string]interface{}{
			"url": payload,
		},
	}
    // Set ToolInputs as well since Execute uses it
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute with timeout because git clone will likely hang or fail slow
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

	_, err := localTool.Execute(ctx, req)

    // Check if file was created
    if _, statErr := os.Stat(tmpFile); statErr == nil {
		t.Errorf("VULNERABILITY CONFIRMED: RCE via git ext:: protocol. File %s created.", tmpFile)
	} else {
        // Safe, now verify it was blocked by our new check
        if err == nil {
            t.Errorf("Expected error but got nil. Vulnerability might be mitigated by environment but not by our code.")
        } else if !strings.Contains(err.Error(), "git injection detected") {
             // If it failed for another reason (like transport not allowed), it's safe but not because of our fix?
             // Actually, if our fix works, it MUST fail with "git injection detected" BEFORE invoking git.
             t.Errorf("Expected 'git injection detected' error, got: %v", err)
        } else {
             t.Logf("SUCCESS: Attack blocked by Sentinel Security check: %v", err)
        }
	}

    // Test Case 2: Whitespace Bypass Attempt
    t.Run("WhitespaceBypass", func(t *testing.T) {
        bypassPayload := " " + payload
        reqBypass := &ExecutionRequest{
            ToolName: "git-clone",
            Arguments: map[string]interface{}{
                "url": bypassPayload,
            },
        }
        reqBypass.ToolInputs, _ = json.Marshal(reqBypass.Arguments)
        ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
        defer cancel2()

        _, errBypass := localTool.Execute(ctx2, reqBypass)

        if errBypass == nil {
             t.Errorf("Vulnerability: Leading space bypassed security check!")
        } else if !strings.Contains(errBypass.Error(), "git injection detected") {
             t.Errorf("Leading space check failed to catch git injection. Got: %v", errBypass)
        } else {
             t.Logf("SUCCESS: Leading space bypass attempt blocked: %v", errBypass)
        }
    })
}
