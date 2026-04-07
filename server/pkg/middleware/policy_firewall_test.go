// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestPolicyFirewallMiddleware(t *testing.T) {
	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	tests := []struct {
		name          string
		config        PolicyFirewallConfig
		req           *tool.ExecutionRequest
		expectedError string
	}{
		{
			name: "Disabled firewall allows all",
			config: PolicyFirewallConfig{
				Enabled: false,
			},
			req: &tool.ExecutionRequest{ToolName: "test.tool"},
		},
		{
			name: "Explicitly blocked tool is denied",
			config: PolicyFirewallConfig{
				Enabled:      true,
				BlockedTools: []string{"dangerous.tool"},
			},
			req:           &tool.ExecutionRequest{ToolName: "dangerous.tool"},
			expectedError: "policy firewall: access denied for tool 'dangerous.tool' (blocked)",
		},
		{
			name: "Blocked prefix tool is denied",
			config: PolicyFirewallConfig{
				Enabled:      true,
				BlockedTools: []string{"aws.*"},
			},
			req:           &tool.ExecutionRequest{ToolName: "aws.delete_bucket"},
			expectedError: "policy firewall: access denied for tool 'aws.delete_bucket' (blocked)",
		},
		{
			name: "Explicitly allowed tool is permitted",
			config: PolicyFirewallConfig{
				Enabled:      true,
				AllowedTools: []string{"safe.tool"},
				DefaultAction: "deny",
			},
			req: &tool.ExecutionRequest{ToolName: "safe.tool"},
		},
		{
			name: "Allowed prefix tool is permitted",
			config: PolicyFirewallConfig{
				Enabled:      true,
				AllowedTools: []string{"fs.read.*"},
				DefaultAction: "deny",
			},
			req: &tool.ExecutionRequest{ToolName: "fs.read.file"},
		},
		{
			name: "Default deny blocks unknown tools",
			config: PolicyFirewallConfig{
				Enabled:      true,
				AllowedTools: []string{"safe.tool"},
				DefaultAction: "deny",
			},
			req:           &tool.ExecutionRequest{ToolName: "unknown.tool"},
			expectedError: "policy firewall: access denied for tool 'unknown.tool' (default deny)",
		},
		{
			name: "Default allow permits unknown tools",
			config: PolicyFirewallConfig{
				Enabled:      true,
				AllowedTools: []string{"safe.tool"},
				DefaultAction: "allow",
			},
			req: &tool.ExecutionRequest{ToolName: "unknown.tool"},
		},
		{
			name: "Blocklist overrides allowlist",
			config: PolicyFirewallConfig{
				Enabled:      true,
				AllowedTools: []string{"fs.*"},
				BlockedTools: []string{"fs.delete"},
				DefaultAction: "deny",
			},
			req:           &tool.ExecutionRequest{ToolName: "fs.delete"},
			expectedError: "policy firewall: access denied for tool 'fs.delete' (blocked)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewPolicyFirewallMiddleware(tt.config)
			res, err := m.Execute(context.Background(), tt.req, mockNext)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "success", res)
			}
		})
	}
}
