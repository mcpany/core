// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MockTool is a mock implementation of the Tool interface for testing purposes.
//
// Summary: Mock tool for testing.
type MockTool struct {
	ToolFunc           func() *v1.Tool
	MCPToolFunc        func() *mcp.Tool
	ExecuteFunc        func(ctx context.Context, req *ExecutionRequest) (any, error)
	GetCacheConfigFunc func() *configv1.CacheConfig
}

// Tool provides tool functionality.
//
// Summary: Tool.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (m *MockTool) Tool() *v1.Tool {
	if m.ToolFunc != nil {
		return m.ToolFunc()
	}
	return &v1.Tool{}
}

// MCPTool provides mcptool functionality.
//
// Summary: MCPTool.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (m *MockTool) MCPTool() *mcp.Tool {
	if m.MCPToolFunc != nil {
		return m.MCPToolFunc()
	}
	return nil
}

// Execute provides execute functionality.
//
// Summary: Execute.
//
// Parameters.
//   - ctx: The parameter.
//   - req: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (m *MockTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// GetCacheConfig provides getcacheconfig functionality.
//
// Summary: GetCacheConfig.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	if m.GetCacheConfigFunc != nil {
		return m.GetCacheConfigFunc()
	}
	return nil
}
