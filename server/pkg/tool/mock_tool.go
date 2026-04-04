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
// Summary: Is a mock implementation of the Tool interface for testing purposes.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

type MockTool struct {
	ToolFunc           func() *v1.Tool
	MCPToolFunc        func() *mcp.Tool
	ExecuteFunc        func(ctx context.Context, req *ExecutionRequest) (any, error)
	GetCacheConfigFunc func() *configv1.CacheConfig
}

// Tool returns the protobuf definition of the mock tool.
//
// Summary: Returns the protobuf definition of the mock tool.
//
// Parameters:
//   - None.
//
// Returns:
//   - *v1.Tool: Return value.
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

// MCPTool returns the MCP tool definition.
//
// Summary: Returns the MCP tool definition.
//
// Parameters:
//   - None.
//
// Returns:
//   - *mcp.Tool: Return value.
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
//
// Errors:
//   - Returns the error returned by the underlying mock ExecuteFunc.
//
// Side Effects:
//   - Calls the underlying mock ExecuteFunc.
func (m *MockTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// GetCacheConfig calls the mock GetCacheConfigFunc if set, otherwise returns nil.
//
// Summary: Calls the mock GetCacheConfigFunc if set, otherwise returns nil.
//
// Parameters:
//   - None.
//
// Returns:
//   - *configv1.CacheConfig: Return value.
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
