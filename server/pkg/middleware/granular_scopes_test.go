// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestGranularScopesMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		config      GranularScopesConfig
		toolName    string
		expectError bool
	}{
		{
			name: "Explicitly Allowed",
			config: GranularScopesConfig{
				Default: "deny",
				Tokens:  []string{"fs:read:/tmp"},
			},
			toolName:    "fs.read",
			expectError: false, // "fs:read:/tmp" starts with "fs:read"
		},
		{
			name: "Explicitly Denied",
			config: GranularScopesConfig{
				Default: "deny",
				Tokens:  []string{"fs:read:/tmp"},
			},
			toolName:    "db.write",
			expectError: true,
		},
		{
			name: "Default Read Allows Read Tool",
			config: GranularScopesConfig{
				Default: "read",
				Tokens:  []string{},
			},
			toolName:    "fs.read",
			expectError: false,
		},
		{
			name: "Default Read Denies Write Tool",
			config: GranularScopesConfig{
				Default: "read",
				Tokens:  []string{},
			},
			toolName:    "fs.write",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewGranularScopesMiddleware(tt.config)
			req := &tool.ExecutionRequest{ToolName: tt.toolName}

			called := false
			next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
				called = true
				return "success", nil
			}

			_, err := m.Execute(context.Background(), req, next)

			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, called)
			} else {
				assert.NoError(t, err)
				assert.True(t, called)
			}
		})
	}
}
