// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSSRFLoopbackShorthand(t *testing.T) {
	// Setup helper to create tool
	createTool := func(cmd string) Tool {
		toolDef := v1.Tool_builder{Name: proto.String("test-tool")}.Build()
		service := configv1.CommandLineUpstreamService_builder{
			Command: &cmd,
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"{{arg}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("arg")}.Build(),
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
			name:          "Block 127.1 shorthand",
			input:         "127.1",
			expectError:   true,
			errorContains: "loopback shorthand address is not allowed",
		},
		{
			name:          "Block 127.0.1 shorthand",
			input:         "127.0.1",
			expectError:   true,
			errorContains: "loopback shorthand address is not allowed",
		},
		{
			name:          "Block 127.255 shorthand",
			input:         "127.255",
			expectError:   true,
			errorContains: "loopback shorthand address is not allowed",
		},
		{
			name:          "Block 127.0.0.1 (redundant but checked)",
			input:         "127.0.0.1",
			expectError:   true,
			errorContains: "loopback shorthand address is not allowed", // Could be IsSafeIP or our new check depending on mocks
		},
		{
			name:          "Allow 127.txt (filename)",
			input:         "127.txt",
			expectError:   false,
		},
		{
			name:          "Allow 127-1 (filename)",
			input:         "127-1",
			expectError:   false,
		},
		{
			name:          "Allow 10.1 (private but not loopback shorthand)",
			input:         "10.1",
			expectError:   false, // Assuming mock allows private IPs or net.ParseIP fails
		},
		{
			name:          "Allow 0 (ambiguous)",
			input:         "0",
			expectError:   false,
		},
		{
			name:          "Allow 127.1 if loopback allowed",
			input:         "127.1",
			allowLoopback: true,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock IsSafeIP to bypass standard IP checks so we isolate our new check
			// (or just rely on net.ParseIP failing for shorthands)
			originalIsSafeIP := validation.IsSafeIP
			validation.IsSafeIP = func(ipStr string) error {
				// Simulate net.ParseIP failure for shorthands, success for standard IPs
				ip := net.ParseIP(ipStr)
				if ip == nil {
					return fmt.Errorf("invalid IP address")
				}
				// If loopback allowed, everything passes IsSafeIP
				// If not, we might block 127.0.0.1 here.
				// But we want to test the shorthand check primarily.
				return nil
			}
			defer func() { validation.IsSafeIP = originalIsSafeIP }()

			if tt.allowLoopback {
				os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
			} else {
				os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
			}
			defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

			tool := createTool("echo")
			req := &ExecutionRequest{
				ToolName:   "test",
				ToolInputs: []byte(`{"arg": "` + tt.input + `"}`),
				DryRun:     true,
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
