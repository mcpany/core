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

func TestSSRFNetworkToolProtection(t *testing.T) {
	// Unset dangerous bypass to test actual validation logic
	os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
	defer os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// Mock validation.LookupIP
	originalLookupIP := validation.LookupIP
	defer func() { validation.LookupIP = originalLookupIP }()

	validation.LookupIP = func(ctx context.Context, network, host string) ([]net.IP, error) {
		switch host {
		case "localtest.me":
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		case "private.internal":
			return []net.IP{net.ParseIP("10.0.0.1")}, nil
		case "google.com":
			return []net.IP{net.ParseIP("8.8.8.8")}, nil
		case "file.txt":
			return nil, fmt.Errorf("no such host")
		default:
			return nil, fmt.Errorf("unknown host")
		}
	}

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
		command       string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name:          "Block schema-less loopback (curl)",
			command:       "curl",
			input:         "localtest.me",
			expectError:   true,
			errorContains: "unsafe network argument",
		},
		{
			name:          "Block schema-less private IP (wget)",
			command:       "wget",
			input:         "private.internal",
			expectError:   true,
			errorContains: "unsafe network argument",
		},
		{
			name:          "Block @ LFI (curl)",
			command:       "curl",
			input:         "@/etc/passwd",
			expectError:   true,
			errorContains: "'@' prefix is not allowed",
		},
		{
			name:        "Allow public domain (curl)",
			command:     "curl",
			input:       "google.com",
			expectError: false,
		},
		{
			name:        "Allow file path (resolution fails)",
			command:     "curl",
			input:       "file.txt",
			expectError: false,
		},
		{
			name:        "Allow non-network tool (echo)",
			command:     "echo",
			input:       "localtest.me",
			expectError: false,
		},
		{
			name:          "Block existing checks (localhost)",
			command:       "curl",
			input:         "localhost",
			expectError:   true,
			errorContains: "localhost is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := createTool(tt.command)
			req := &ExecutionRequest{
				ToolName:   "test",
				ToolInputs: []byte(`{"arg": "` + tt.input + `"}`),
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
