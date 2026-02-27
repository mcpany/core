// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSSRFArgumentProtection_HostPort(t *testing.T) {
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
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name:        "Block host:port (localhost)",
			input:       "localhost:8080",
			expectError: true,
			errorContains: "localhost is not allowed",
		},
		{
			name:        "Block host:port (127.0.0.1)",
			input:       "127.0.0.1:8080",
			expectError: true,
			errorContains: "loopback address is not allowed",
		},
		{
			name:        "Block host:port (private IP)",
			input:       "192.168.1.1:22",
			expectError: true,
			errorContains: "private network address is not allowed",
		},
		{
			name:        "Block host:port (IPv6 localhost)",
			input:       "[::1]:80",
			expectError: true,
			errorContains: "loopback address is not allowed",
		},
        {
			name:        "Block host:port (metadata service)",
			input:       "169.254.169.254:80",
			expectError: true,
			errorContains: "link-local address is not allowed",
		},
		{
			name:        "Allow host:port (public IP)",
			input:       "8.8.8.8:53",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
            // Use t.Setenv to safely control environment variables for validation.IsSafeIP
            t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")
            t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "false")
            t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "false")

			tool := createTool("curl")
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
