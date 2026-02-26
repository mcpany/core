// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSSRFArgumentProtection(t *testing.T) {
	// Note: We avoid mocking validation.IsSafeURL or validation.IsSafeIP here to prevent
	// DATA RACE with parallel tests (e.g. TestCommandTool_Execute) that might read
	// those global function variables.
	// Instead, we rely on the real implementation which uses net.LookupIP and environment variables.
	// We must ensure environment variables are managed safely (though os.Setenv is also risky in parallel tests).
	// Ideally, this test should run sequentially (no t.Parallel()), which is the case.
	// We hope other tests don't read these env vars concurrently.

	// Setup helper to create tool
	createTool := func(cmd string) Tool {
		toolDef := v1.Tool_builder{Name: proto.String("test-tool")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"{{url}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("url")}.Build(),
				}.Build(),
			},
		}.Build()
		return NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
	}

	tests := []struct {
		name          string
		command       string
		input         string
		allowLoopback bool
		allowPrivate  bool
		expectError   bool
		errorContains string
	}{
		{
			name:          "Block localhost",
			command:       "curl",
			input:         "localhost",
			expectError:   true,
			errorContains: "localhost is not allowed",
		},
		{
			name:          "Block loopback IP",
			command:       "curl",
			input:         "127.0.0.1",
			expectError:   true,
			errorContains: "loopback address is not allowed",
		},
		{
			name:          "Block private IP",
			command:       "wget",
			input:         "192.168.1.1",
			expectError:   true,
			errorContains: "private network address is not allowed",
		},
		{
			name:          "Block metadata service IP",
			command:       "fetch",
			input:         "169.254.169.254",
			expectError:   true,
			errorContains: "link-local address is not allowed",
		},
		{
			name:        "Allow public IP",
			command:     "curl",
			input:       "8.8.8.8",
			expectError: false,
		},
		{
			name:          "Allow localhost if enabled",
			command:       "curl",
			input:         "localhost",
			allowLoopback: true,
			expectError:   false,
		},
		{
			name:          "Allow loopback IP if enabled",
			command:       "curl",
			input:         "127.0.0.1",
			allowLoopback: true,
			expectError:   false,
		},
		{
			name:         "Allow private IP if enabled",
			command:      "curl",
			input:        "10.0.0.1",
			allowPrivate: true,
			expectError:  false,
		},
		{
			name:        "Allow normal filename",
			command:     "cat",
			input:       "myfile.txt",
			expectError: false,
		},
		{
			name:          "Block unsafe URL with scheme (existing check)",
			command:       "curl",
			input:         "http://127.0.0.1",
			expectError:   true,
			errorContains: "unsafe url argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for the test case
			if tt.allowLoopback {
				os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
			} else {
				os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
			}
			if tt.allowPrivate {
				os.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")
			} else {
				os.Unsetenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES")
			}
			// Ensure DANGEROUS flag is unset for this test to verify blocking logic
			os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")

			defer func() {
				// Cleanup
				os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
				os.Unsetenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES")
				os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
			}()

			tool := createTool(tt.command)
			req := &ExecutionRequest{
				ToolName:   "test",
				ToolInputs: []byte(`{"url": "` + tt.input + `"}`),
				DryRun:     true, // Use DryRun to avoid actual execution but trigger validation
			}

			_, err := tool.Execute(context.Background(), req)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
