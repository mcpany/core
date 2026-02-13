// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
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
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("GIT_ALLOW_PROTOCOL"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// The payload: ext::echo PWNED
	payload := "ext::echo PWNED"

	req := &ExecutionRequest{
		ToolName: "git-clone",
		Arguments: map[string]interface{}{
			"url": payload,
            "GIT_ALLOW_PROTOCOL": "ext:ssh:git:http:https:file",
		},
	}
    // Set ToolInputs as well since Execute uses it
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute with timeout because git clone will likely hang or fail slow
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

	res, err := localTool.Execute(ctx, req)

    output := ""
    if err != nil {
        output += err.Error()
    }
    if res != nil {
         if m, ok := res.(map[string]interface{}); ok {
             if combined, ok := m["combined_output"].(string); ok {
                 output += combined
             }
         }
    }

    t.Logf("Output: %s", output)

    // Check if output contains PWNED
    // NOTE: The error message might contain "PWNED" because it echoes the input.
    // We need to check if the error is about "ext: scheme detected" which means it was blocked.
    if strings.Contains(output, "ext: scheme detected") {
        t.Logf("SUCCESS: RCE attempt blocked.")
    } else if strings.Contains(output, "bad line length character") {
		t.Errorf("VULNERABILITY CONFIRMED: RCE via git ext:: protocol. Output contains execution artifacts.")
    } else if strings.Contains(output, "PWNED") && !strings.Contains(output, "ext: scheme detected") {
        // If PWNED is present but NOT part of the blocking message, it might be RCE.
        // But "bad line length character" is the surest sign of git helper execution.
        t.Errorf("VULNERABILITY POTENTIAL: Output contains 'PWNED' but not blocking message.")
	} else {
        // If it failed for other reasons (e.g. timeout), it might be safe or just broken test.
        // But if we blocked it, we are good.
        t.Errorf("Unexpected output: %s", output)
	}
}
