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

// MockTool mockTool represents a mock tool.
//
// Summary: MockTool represents a mock tool.
type MockTool struct {
	ExecuteFunc func(ctx context.Context, req *tool.ExecutionRequest) (any, error)
}

// Tool tool tool.
//
// Summary: Tool tool.
//
// Parameters:
//   - None.
//
// Returns:
//   - *v1.Tool: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *MockTool) Tool() *v1.Tool {
	return v1.Tool_builder{
		Name: proto.String("mock-tool"),
	}.Build()
}

// Execute calls the mock ExecuteFunc if set, otherwise returns nil.
//
// Summary: Executes the mock tool logic.
//
// Parameters: - None.
//   - ctx: context.Context. The execution context.
//   - req: *tool.ExecutionRequest. The tool execution request.
//
// Returns: - None.
//   - any: The result from ExecuteFunc.
//   - error: The error from ExecuteFunc.
//
// Side Effects: - None.
//   - Invokes the injected ExecuteFunc.
func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// GetCacheConfig retrieves the cache config.
//
// Summary: Retrieves the cache config.
//
// Parameters:
//   - None.
//
// Returns:
//   - *configv1.CacheConfig: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
