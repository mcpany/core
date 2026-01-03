// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package testutil provides test utilities and mocks.
package testutil

import (
	"context"

	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
)

// MockTool is a mock implementation of the tool.Tool interface for testing.
type MockTool struct {
	ExecuteFunc func(ctx context.Context, req *tool.ExecutionRequest) (any, error)
}

// Tool returns a basic tool definition for the mock tool.
func (m *MockTool) Tool() *v1.Tool {
	name := "mock-tool"
	return v1.Tool_builder{Name: &name}.Build()
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
