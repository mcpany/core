// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package testutil provides test utilities and mocks.
package testutil

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"google.golang.org/protobuf/proto"
)

// MockTool is a mock implementation of the tool.Tool interface for testing.
//
// Summary: is a mock implementation of the tool.Tool interface for testing.
type MockTool struct {
	ExecuteFunc func(ctx context.Context, req *tool.ExecutionRequest) (any, error)
}

// Tool returns a basic tool definition for the mock tool.
//
// Summary: returns a basic tool definition for the mock tool.
//
// Parameters:
//   None.
//
// Returns:
//   - *v1.Tool: The *v1.Tool.
func (m *MockTool) Tool() *v1.Tool {
	return v1.Tool_builder{
		Name: proto.String("mock-tool"),
	}.Build()
}

// Execute calls the mock ExecuteFunc if set, otherwise returns nil.
//
// Summary: calls the mock ExecuteFunc if set, otherwise returns nil.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - req: *tool.ExecutionRequest. The req.
//
// Returns:
//   - any: The any.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// GetCacheConfig returns nil for the mock tool.
//
// Summary: returns nil for the mock tool.
//
// Parameters:
//   None.
//
// Returns:
//   - *configv1.CacheConfig: The *configv1.CacheConfig.
func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
