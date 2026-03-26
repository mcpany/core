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

// Tool returns the protobuf definition of the mock tool.
//
// Summary: Retrieves the mock tool definition.
//
// Returns:
//   - *v1.Tool: The tool definition.
func (m *MockTool) Tool() *v1.Tool {
	if m.ToolFunc != nil {
		return m.ToolFunc()
	}
	return &v1.Tool{}
}

// MCPTool returns the MCP tool definition.
//
// Summary: Retrieves the MCP tool definition.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
func (m *MockTool) MCPTool() *mcp.Tool {
	if m.MCPToolFunc != nil {
		return m.MCPToolFunc()
	}
	return nil
}

<<<<<<< HEAD
// IsStreaming returns whether the mock tool supports streaming.
//
// Summary: Checks if the mock tool supports streaming.
//
// Parameters:
//   - None.
//
// Returns:
//   - bool: Always returns false.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *MockTool) IsStreaming() bool {
	return false
}

// StreamExecute executes the mock tool in streaming mode.
//
// Summary: Executes the mock tool in streaming mode.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *ExecutionRequest. The execution request.
//
// Returns:
//   - <-chan any: A channel that emits the result or error.
//   - error: Always nil for the mock tool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Executes the mock tool logic asynchronously.
func (m *MockTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
	ch := make(chan any, 1)
	go func() {
		defer close(ch)
		res, err := m.Execute(ctx, req)
		if err != nil {
			ch <- err
		} else {
			ch <- res
		}
	}()
	return ch, nil
}

=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
// Execute calls the mock ExecuteFunc if set, otherwise returns nil.
//
// Summary: Executes the mock tool.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *ExecutionRequest. The execution request.
//
// Returns:
//   - any: The execution result.
//   - error: An error if execution fails.
<<<<<<< HEAD
//
// Errors:
//   - Returns the error returned by the underlying mock ExecuteFunc.
//
// Side Effects:
//   - Calls the underlying mock ExecuteFunc.
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
func (m *MockTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// GetCacheConfig calls the mock GetCacheConfigFunc if set, otherwise returns nil.
//
// Summary: Retrieves the cache configuration.
//
// Returns:
//   - *configv1.CacheConfig: The cache configuration.
func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	if m.GetCacheConfigFunc != nil {
		return m.GetCacheConfigFunc()
	}
	return nil
}
