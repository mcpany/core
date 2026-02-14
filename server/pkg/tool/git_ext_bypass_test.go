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

func TestLocalCommandTool_Git_Ext_Bypass(t *testing.T) {
	// This test confirms that the protection against "ext::" is case-insensitive.
    // It attempts to bypass using "EXT::".

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

	// The payload with uppercase EXT::
	payload := "EXT::sh -c echo_pwned"

	req := &ExecutionRequest{
		ToolName: "git-clone",
		Arguments: map[string]interface{}{
			"url": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := localTool.Execute(ctx, req)

	if err != nil {
		if strings.Contains(err.Error(), "git ext:: protocol is not allowed") {
			// Passed: Blocked as expected
		} else {
			// It failed for some other reason, meaning check was bypassed (if it was blocked, error message would match).
            // However, since we haven't fixed it yet, this will likely fail with "executable not found" or similar if git runs.
            // But for the FINAL test (after fix), we expect the specific error.
			t.Errorf("Bypassed or unexpected error: %v", err)
		}
	} else {
		// If it succeeded, it definitely bypassed the check.
		t.Error("Bypassed: Execution proceeded without error.")
	}
}
