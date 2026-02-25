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

func TestSSRFLoopbackShorthandProtection(t *testing.T) {
	// Setup helper to create tool
	createTool := func() Tool {
		cmd := "curl"
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
		allowLoopback bool
		expectError   bool
		errorContains string
	}{
		{
			name:          "Block 127.1",
			input:         "127.1",
			expectError:   true,
			errorContains: "loopback shorthand is not allowed",
		},
		{
			name:          "Block 127.0.1",
			input:         "127.0.1",
			expectError:   true,
			errorContains: "loopback shorthand is not allowed",
		},
		{
			name:          "Block 127.255",
			input:         "127.255",
			expectError:   true,
			errorContains: "loopback shorthand is not allowed",
		},
		{
			name:          "Block 0177.1 (Octal)",
			input:         "0177.1",
			expectError:   true,
			errorContains: "loopback shorthand is not allowed",
		},
		{
			name:          "Allow 127.1 if allowLoopback=true",
			input:         "127.1",
			allowLoopback: true,
			expectError:   false,
		},
		{
			name:        "Allow 127.txt (filename)",
			input:       "127.txt",
			expectError: false,
		},
		{
			name:        "Allow 127.1.txt (filename)",
			input:       "127.1.txt",
			expectError: false,
		},
		{
			name:        "Allow 10.1 (Ambiguous/Version/Private IP) - currently allowed by logic",
			input:       "10.1",
			expectError: false,
		},
		{
			name:        "Allow 0 (Ambiguous) - currently allowed by logic",
			input:       "0",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.allowLoopback {
				os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
			} else {
				os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
			}
			defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

			tool := createTool()
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
