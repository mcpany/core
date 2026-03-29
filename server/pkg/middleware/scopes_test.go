// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"
	"encoding/json"

	"github.com/mcpany/core/server/pkg/tool"
)

func TestScopesMiddleware(t *testing.T) {
	config := ScopesConfig{
		Roles: map[string][]string{
			"default": {"fs:read:/tmp"},
			"admin":   {"fs:", "db:"},
		},
	}
	middleware := NewScopesMiddleware(config)

	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	tests := []struct {
		name        string
		role        string
		toolName    string
		toolInputs  json.RawMessage
		expectError bool
	}{
		{
			name:        "default role allowed",
			role:        "default",
			toolName:    "fs:read",
			toolInputs:  json.RawMessage(`{"path": "/tmp/file.txt"}`),
			expectError: false,
		},
		{
			name:        "default role denied",
			role:        "default",
			toolName:    "fs:write",
			toolInputs:  json.RawMessage(`{"path": "/tmp/file.txt"}`),
			expectError: true,
		},
		{
			name:        "admin role allowed fs",
			role:        "admin",
			toolName:    "fs:write",
			toolInputs:  json.RawMessage(`{"path": "/etc/passwd"}`),
			expectError: false,
		},
		{
			name:        "admin role allowed db",
			role:        "admin",
			toolName:    "db:drop",
			toolInputs:  json.RawMessage(`{"table": "users"}`),
			expectError: false,
		},
		{
			name:        "admin role denied network",
			role:        "admin",
			toolName:    "network:connect",
			toolInputs:  json.RawMessage(`{"host": "localhost"}`),
			expectError: true,
		},
		{
			name:        "unknown role",
			role:        "unknown",
			toolName:    "fs:read",
			toolInputs:  json.RawMessage(`{"path": "/tmp/file.txt"}`),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), agentRoleKey, tt.role)
			req := &tool.ExecutionRequest{ToolName: tt.toolName, ToolInputs: tt.toolInputs}

			_, err := middleware.Execute(ctx, req, mockNext)

			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
