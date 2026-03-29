// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"


	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestDiscoverySandboxMiddleware_PreExecute(t *testing.T) {
	tests := []struct {
		name        string
		config      DiscoverySandboxConfig
		toolName    string
		expectError bool
	}{
		{
			name: "Enabled and Configured for Discovery",
			config: DiscoverySandboxConfig{
				Enabled:             true,
				IsolatedEnvironment: "gVisor",
				MaxExecutionTimeMs:  500,
			},
			toolName:    "get_discovery_manifest",
			expectError: false,
		},
		{
			name: "Enabled Missing Environment",
			config: DiscoverySandboxConfig{
				Enabled:            true,
				MaxExecutionTimeMs: 500,
			},
			toolName:    "system_discovery",
			expectError: true,
		},
		{
			name: "Enabled Missing Timeout",
			config: DiscoverySandboxConfig{
				Enabled:             true,
				IsolatedEnvironment: "gVisor",
			},
			toolName:    "system_discovery",
			expectError: true,
		},
		{
			name: "Non-Discovery Command",
			config: DiscoverySandboxConfig{
				Enabled:             true,
				IsolatedEnvironment: "gVisor",
				MaxExecutionTimeMs:  500,
			},
			toolName:    "read_file",
			expectError: false,
		},
		{
			name: "Disabled",
			config: DiscoverySandboxConfig{
				Enabled: false,
			},
			toolName:    "system_discovery",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewDiscoverySandboxMiddleware(tt.config)

			req := &mcp.CallToolRequest{
				Params: &mcp.CallToolParamsRaw{
					Name: tt.toolName,
				},
			}

			err := m.PreExecute(context.Background(), req, nil)

			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDiscoverySandboxMiddleware_PostExecute(t *testing.T) {
	config := DiscoverySandboxConfig{
		Enabled:             true,
		IsolatedEnvironment: "gVisor",
		MaxExecutionTimeMs:  500,
	}
	m := NewDiscoverySandboxMiddleware(config)

	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name: "discovery_command",
		},
	}

	res, err := m.PostExecute(context.Background(), req, &mcp.CallToolResult{}, nil)
	if err != nil {
		t.Errorf("unexpected error from PostExecute: %v", err)
	}
	if res == nil {
		t.Errorf("expected result, got nil")
	}
}
