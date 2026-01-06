// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package testutil provides test utilities and mocks.
package testutil

import (
	"context"

	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MockTool is a mock implementation of the tool.Tool interface for testing.
type MockTool struct {
	ToolDef     *v1.Tool
	ExecuteFunc func(ctx context.Context, req *tool.ExecutionRequest) (any, error)
}

// Tool returns the tool definition for the mock tool.
func (m *MockTool) Tool() *v1.Tool {
	if m.ToolDef != nil {
		return m.ToolDef
	}
	name := "mock-tool"
	return &v1.Tool{Name: &name}
}

// MCPTool returns the MCP tool definition.
func (m *MockTool) MCPTool() *mcp.Tool {
	t := m.Tool()
	return &mcp.Tool{
		Name:        t.GetName(),
		Description: t.GetDescription(),
	}
}

// Execute calls the mock ExecuteFunc if set, otherwise returns nil.
func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// GetCacheConfig returns nil for the mock tool.
func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

var _ tool.Tool = (*MockTool)(nil)
