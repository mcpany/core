// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_SSRF_Repro(t *testing.T) {
    // Unset the bypass env var set by TestMain to test the protection
    os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
    // Restore it after test
    defer os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// This test demonstrates that LocalCommandTool allows unsafe URLs (SSRF).
    // We use "echo" as the command to simulate a tool that takes a URL (like curl/wget).
    // If validation was present, it should block the unsafe URL before execution.

	tool := configv1.ToolDefinition_builder{
		Name: proto.String("curl-wrapper"),
	}.Build()

    toolProto := v1.Tool_builder{
        Name: proto.String("curl-wrapper"),
    }.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"), // Simulating curl
		Local:   proto.Bool(true),
        Tools:   []*configv1.ToolDefinition{tool},
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{url}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("url"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// The payload: a local URL that should be blocked
	payload := "http://127.0.0.1:22/secret"

	req := &ExecutionRequest{
		ToolName: "curl-wrapper",
		Arguments: map[string]interface{}{
			"url": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// We expect NO error currently (vulnerability), but checking for error means we want one.
    // If Execute returns nil error, it means the URL was passed to the command.
	_, err := localTool.Execute(context.Background(), req)

    if err == nil {
        t.Errorf("VULNERABILITY CONFIRMED: Allowed unsafe URL %s", payload)
    } else {
        t.Logf("Safe: blocked URL with error: %v", err)
    }
}
