// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MockTool - Auto-generated documentation.
//
// Summary: MockTool is a mock implementation of the Tool interface for testing purposes.
//
// Fields:
//   - Various fields for MockTool.
type MockTool struct {
	ToolFunc           func() *v1.Tool
	MCPToolFunc        func() *mcp.Tool
	ExecuteFunc        func(ctx context.Context, req *ExecutionRequest) (any, error)
	GetCacheConfigFunc func() *configv1.CacheConfig
}

// Tool - Auto-generated documentation.
//
// Summary: Tool returns the protobuf definition of the mock tool.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (m *MockTool) Tool() *v1.Tool {
	if m.ToolFunc != nil {
		return m.ToolFunc()
	}
	return &v1.Tool{}
}

// MCPTool - Auto-generated documentation.
//
// Summary: MCPTool returns the MCP tool definition.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (m *MockTool) MCPTool() *mcp.Tool {
	if m.MCPToolFunc != nil {
		return m.MCPToolFunc()
	}
	return nil
}

// Execute calls the mock ExecuteFunc if set, otherwise returns nil. Summary: Executes the mock tool. Parameters: - ctx: context.Context. The execution context. - req: *ExecutionRequest. The execution request. Returns: - any: The execution result. - error: An error if execution fails.
//
// Summary: Execute calls the mock ExecuteFunc if set, otherwise returns nil. Summary: Executes the mock tool. Parameters: - ctx: context.Context. The execution context. - req: *ExecutionRequest. The execution request. Returns: - any: The execution result. - error: An error if execution fails.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//   - req (*ExecutionRequest): The request object containing specific parameters.
//
// Returns:
//   - (any): The resulting any object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
func (m *MockTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// GetCacheConfig - Auto-generated documentation.
//
// Summary: GetCacheConfig calls the mock GetCacheConfigFunc if set, otherwise returns nil.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	if m.GetCacheConfigFunc != nil {
		return m.GetCacheConfigFunc()
	}
	return nil
}
