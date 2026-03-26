// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MockTool mockTool represents a mock tool.
//
// Summary: MockTool represents a mock tool.
type MockTool struct {
	ToolFunc           func() *v1.Tool
	MCPToolFunc        func() *mcp.Tool
	ExecuteFunc        func(ctx context.Context, req *ExecutionRequest) (any, error)
	GetCacheConfigFunc func() *configv1.CacheConfig
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
	if m.ToolFunc != nil {
		return m.ToolFunc()
	}
	return &v1.Tool{}
}

// MCPTool mCPTool mcp tool.
//
// Summary: MCPTool mcp tool.
//
// Parameters:
//   - None.
//
// Returns:
//   - *mcp.Tool: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *MockTool) MCPTool() *mcp.Tool {
	if m.MCPToolFunc != nil {
		return m.MCPToolFunc()
	}
	return nil
}

// Execute calls the mock ExecuteFunc if set, otherwise returns nil.
//
// Summary: Executes the mock tool.
//
// Parameters: - None.
//   - ctx: context.Context. The execution context.
//   - req: *ExecutionRequest. The execution request.
//
// Returns: - None.
//   - any: The execution result.
//   - error: An error if execution fails.
func (m *MockTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
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
	if m.GetCacheConfigFunc != nil {
		return m.GetCacheConfigFunc()
	}
	return nil
}
