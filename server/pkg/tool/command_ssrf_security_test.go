// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_SSRF_Repro(t *testing.T) {
	// This test demonstrates that LocalCommandTool blocks access to loopback addresses
	// via commands that accept URLs (like git) or arguments containing URLs.

	// 1. Start a local HTTP server
	// Ensure safety checks are enabled
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "false")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false") // Should be false by default/unset, but explicit is safer

	requestReceived := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestReceived)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 2. Configure a tool that runs 'git ls-remote {{url}}'
	toolProto := v1.Tool_builder{
		Name: proto.String("git-ls"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"ls-remote", "{{url}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("url"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Test Cases
	tests := []struct {
		name        string
		input       string
		shouldBlock bool
	}{
		{
			name:        "Direct unsafe URL",
			input:       server.URL,
			shouldBlock: true,
		},
		{
			name:        "Embedded unsafe URL (key=value)",
			input:       "url=" + server.URL,
			shouldBlock: true,
		},
		{
			name:        "Embedded unsafe URL (text)",
			input:       "Check " + server.URL + " for issues",
			shouldBlock: true,
		},
		{
			name:        "Safe URL",
			input:       "https://google.com",
			shouldBlock: false,
		},
		{
			name:        "Embedded Safe URL",
			input:       "--url=https://google.com",
			shouldBlock: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := &ExecutionRequest{
				ToolName: "git-ls",
				Arguments: map[string]interface{}{
					"url": tc.input,
				},
			}
			req.ToolInputs, _ = json.Marshal(req.Arguments)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			_, err := localTool.Execute(ctx, req)

			if tc.shouldBlock {
				if err == nil {
					// Check if request reached server
					select {
					case <-requestReceived:
						t.Errorf("VULNERABILITY: LocalCommandTool allowed connection to unsafe address %s", tc.input)
						// Reset channel for next test
						requestReceived = make(chan struct{})
					case <-time.After(100 * time.Millisecond):
						// Failed but maybe not due to SSRF check?
						// If err is nil, it means Execute succeeded (started command).
						// But if git failed to connect (e.g. invalid repo), it returns result with error code.
						// We expect Execute to return ERROR if blocked by policy.
						t.Errorf("Expected SSRF block error, got nil")
					}
				} else {
					if !strings.Contains(err.Error(), "unsafe url target") {
						t.Logf("Got error but not expected SSRF block: %v", err)
					} else {
						// Success
					}
				}
			} else {
				// Should not block due to SSRF. Might fail due to git execution (google.com not a repo)
				// We just check that error is NOT "unsafe url target"
				if err != nil && strings.Contains(err.Error(), "unsafe url target") {
					t.Errorf("False Positive: Safe URL blocked: %v", err)
				}
			}
		})
	}
}
