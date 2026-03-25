// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
// MockTool is a mock implementation of the Tool interface for testing purposes.
//
// Summary: Mock tool for testing.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Tool returns the protobuf definition of the mock tool.
//
// Summary: Retrieves the mock tool definition.
//
// Returns:
//   - *v1.Tool: The tool definition.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// MCPTool returns the MCP tool definition.
//
// Summary: Retrieves the MCP tool definition.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
// GetCacheConfig calls the mock GetCacheConfigFunc if set, otherwise returns nil.
//
// Summary: Retrieves the cache configuration.
//
// Returns:
//   - *configv1.CacheConfig: The cache configuration.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	if m.GetCacheConfigFunc != nil {
		return m.GetCacheConfigFunc()
	}
	return nil
}
