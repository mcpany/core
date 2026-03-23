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
func (m *MockTool) IsStreaming() bool {
	return false
}

// StreamExecute executes the mock tool in streaming mode.
//
// Summary: Executes the tool in a separate goroutine and returns a channel with the result.
//
// Parameters:
//   - ctx (context.Context): The context for execution.
//   - req (*ExecutionRequest): The request parameters.
//
// Returns:
//   - <-chan any: A channel that will yield the result or error.
//   - error: Returns nil as the channel is returned immediately.
//
// Errors:
//   - None directly, errors are sent to the returned channel.
//
// Side Effects:
//   - Starts a goroutine.
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

// Execute performs the mocked tool execution.
//
// Summary: Records the execution call and returns the configured mocked response.
//
// Parameters:
//   - ctx (context.Context): The context for execution.
//   - req (*ExecutionRequest): The execution parameters.
//
// Returns:
//   - any: The mocked response value.
//   - error: The mocked error, if configured.
//
// Errors:
//   - Returns the configured mocked error if one is set.
//
// Side Effects:
//   - Depends on the configured mock ExecuteFunc.
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
