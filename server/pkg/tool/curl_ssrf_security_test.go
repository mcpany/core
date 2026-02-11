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

func TestLocalCommandTool_Curl_SSRF_Repro(t *testing.T) {
	// Unset the dangerous env var that might be set by TestMain
	os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")

	// This test attempts to demonstrate that curl command execution allows dangerous
	// protocols like gopher:// which can lead to SSRF/RCE on internal services.

	tool := configv1.ToolDefinition_builder{
		Name: proto.String("curl-tool"),
	}.Build()

    toolProto := v1.Tool_builder{
        Name: proto.String("curl-tool"),
    }.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("curl"),
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

	// The payload: gopher://localhost:6379/_SLAVEOF... (Redis attack)
    // We use a dummy gopher URL to verify it's not blocked by validation
	payload := "gopher://localhost:6379/_SLAVEOF"

	req := &ExecutionRequest{
		ToolName: "curl-tool",
		Arguments: map[string]interface{}{
			"url": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	res, err := localTool.Execute(ctx, req)

    // Validation should have blocked this, but it won't.
    // If execution proceeds (even if curl fails to connect), it means validation passed.
    // We check if error is "file: scheme detected" or similar validation error.

    if err != nil {
        errStr := strings.ToLower(err.Error())
        if strings.Contains(errStr, "scheme detected") || strings.Contains(errStr, "unsafe url") || strings.Contains(errStr, "unsafe scheme") {
            t.Logf("Safe: Blocked by validation: %v", err)
            return
        }
        // If it's a curl error (exit code), it means validation passed!
        t.Logf("Execution Error (likely curl failed to connect): %v", err)
    } else {
        t.Logf("Execution Success: %v", res)
    }

    // If we reach here without validation error, it's vulnerable.
    t.Errorf("VULNERABILITY CONFIRMED: SSRF via curl gopher:// protocol not blocked.")
}
