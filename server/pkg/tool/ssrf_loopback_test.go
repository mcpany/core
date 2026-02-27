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
			name:          "Block 127.1 shorthand",
			command:       "curl",
			input:         "127.1",
			expectError:   true,
			errorContains: "loopback address is not allowed",
		},
		{
			name:          "Block 127.0.1 shorthand",
			command:       "curl",
			input:         "127.0.1",
			expectError:   true,
			errorContains: "loopback address is not allowed",
		},
		{
			name:          "Block 127.255 shorthand",
			command:       "curl",
			input:         "127.255",
			expectError:   true,
			errorContains: "loopback address is not allowed",
		},
		{
			name:          "Block 127.0.0.1 (parsed by net.ParseIP)",
			command:       "curl",
			input:         "127.0.0.1",
			expectError:   true,
			errorContains: "loopback address is not allowed",
		},
		{
			name:          "Allow 127.1 if loopback allowed",
			command:       "curl",
			input:         "127.1",
			allowLoopback: true,
			expectError:   false,
		},
		{
			name:          "Block 0177.0.0.1 (octal)",
			command:       "curl",
			input:         "0177.0.0.1",
			expectError:   true,
			errorContains: "loopback address is not allowed",
		},
		{
			name:          "Block 0x7f.1 (hex/dotted)",
			command:       "curl",
			input:         "0x7f.1",
			expectError:   true,
			errorContains: "loopback address is not allowed",
		},
		{
			name:          "Block 10.1 (private shorthand)",
			command:       "curl",
			input:         "10.1",
			expectError:   true,
			errorContains: "private network address is not allowed",
		},
		{
			name:         "Allow 10.1 if private allowed",
			command:      "curl",
			input:        "10.1",
			allowPrivate: true,
			expectError:  false,
		},
		{
			name:          "Allow 127.txt (file)",
			command:       "curl",
			input:         "127.txt", // contains letters (t, x), not hex/octal valid for IP
			expectError:   false,
		},
		{
			name:          "Block 0 (integer) for network tool",
			command:       "curl",
			input:         "0",
			expectError:   true,
			errorContains: "unsafe IP argument (permissive)", // blocked by unspecific address (0.0.0.0) check in ValidateIP
		},
		{
			name:          "Allow 0 (integer) for non-network tool",
			command:       "echo",
			input:         "0",
			expectError:   false,
		},
		{
			name:          "Block 2130706433 (integer) for network tool",
			command:       "wget",
			input:         "2130706433",
			expectError:   true,
			errorContains: "loopback address is not allowed",
		},
		{
			name:          "Allow 3600 (integer) for non-network tool",
			command:       "sleep",
			input:         "3600",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure strict mode for the test
			originalEnvDangerous := os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
			os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
			defer func() {
				if originalEnvDangerous != "" {
					os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", originalEnvDangerous)
				}
			}()

			// Mock IsSafeIP to respect test case flags
			originalIsSafeIP := validation.IsSafeIP
			validation.IsSafeIP = func(ipStr string) error {
				ip := net.ParseIP(ipStr)
				if ip == nil {
					return fmt.Errorf("invalid IP address")
				}
				return validation.ValidateIP(ip, tt.allowLoopback, tt.allowPrivate)
			}
			defer func() { validation.IsSafeIP = originalIsSafeIP }()

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
			defer func() {
				os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
				os.Unsetenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES")
			}()

			tool := createTool(tt.command)
			req := &ExecutionRequest{
				ToolName:   "test",
				ToolInputs: []byte(`{"url": "` + tt.input + `"}`),
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
