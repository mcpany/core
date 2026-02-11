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

func TestLocalCommandTool_Git_Ext_Security(t *testing.T) {
	// This test attempts to demonstrate that git command execution allows injecting
	// dangerous protocols like ext:: which can lead to RCE.
    // We assume the attacker can also set GIT_ALLOW_PROTOCOL via environment variable injection
    // which is not currently blocked by isDangerousEnvVar.

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
            // Allow user to set GIT_ALLOW_PROTOCOL
            configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("GIT_ALLOW_PROTOCOL"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// The payload: ext::sh -c "touch /tmp/pwned_git_ext"
    // We use a harmless command but verify execution via file creation
	payload := "ext::id"

	req := &ExecutionRequest{
		ToolName: "git-clone",
		Arguments: map[string]interface{}{
			"url": payload,
            "GIT_ALLOW_PROTOCOL": "ext:ssh:file:http:https",
		},
	}
    // Set ToolInputs as well since Execute uses it
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute with timeout because git clone will likely hang or fail slow
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

	_, err := localTool.Execute(ctx, req)

    if err == nil {
        t.Errorf("Expected error but got nil. Vulnerability might be mitigated silently or still present.")
    } else {
        if strings.Contains(err.Error(), "ext: scheme detected") {
            t.Logf("Success: RCE attempt blocked by input validation: %v", err)
        } else {
             t.Errorf("Expected 'ext: scheme detected' error, got: %v", err)
        }
    }
}
