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
// IsStreaming returns whether the tool supports streaming execution.
//
// Summary: Returns whether streaming is supported.
//
// Parameters:
//   - None.
//
// Returns:
//   - bool: True if streaming is supported, otherwise false.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *MockTool) IsStreaming() bool {
	return false
}

// StreamExecute simulates streaming execution for testing.
//
// Summary: Simulates streaming execution.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*ExecutionRequest): The execution request.
//
// Returns:
//   - <-chan any: A mock streaming channel.
//   - error: An error if configured to fail.
//
// Errors:
//   - Returns an error if mock fails.
//
// Side Effects:
//   - None.
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

// Execute runs the tool with the given request.
//
// Summary: Executes the tool.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*ExecutionRequest): The execution request parameters.
//
// Returns:
//   - any: The execution result.
//   - error: An error if execution fails.
//
// Errors:
//   - Returns an error on operational failures.
//
// Side Effects:
//   - May invoke upstream services or mutate state depending on the tool logic.
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
// GetCacheConfig retrieves the cache configuration for the tool.
//
// Summary: Retrieves the cache configuration.
//
// Parameters:
//   - None.
//
// Returns:
//   - *configv1.CacheConfig: The configuration or nil.
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
