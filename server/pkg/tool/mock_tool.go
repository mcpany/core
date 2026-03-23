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

// StreamExecute intercepts a mock request to emulate the channel-driven asynchronous flow typically used by streaming tools, injecting predefined events into a buffer.
//
// Summary: Emulates an asynchronous streaming tool by dispatching synthetic mock results to a channel over time.
//
// Parameters:
//   - ctx: context.Context. The parent context enforcing execution deadlines and testing scope cancellation.
//   - req: *ExecutionRequest. The synthetic parameters dispatched to the mocked method pipeline.
//
// Returns:
//   - <-chan any: A read-only testing channel where hardcoded sequences of responses are pushed.
//   - error: Returns the predetermined mock error immediately to simulate an abrupt failure.
//
// Side Effects:
//   - Fires a goroutine filling the mock channel independent of caller execution.
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

// Execute triggers the mock testing pipeline, returning the explicitly configured return payload and appending the invocation to an internal tracker array.
//
// Summary: Simulates a synchronous tool invocation to validate client integration.
//
// Parameters:
//   - ctx: context.Context. The context limiting the simulation lifespan and enforcing test state bounds.
//   - req: *ExecutionRequest. The captured testing payload simulating arbitrary runtime inputs.
//
// Returns:
//   - any: The deterministic object or map explicitly provided during testing setup.
//   - error: Returns a mock error to test downstream exception handling mechanisms.
//
// Errors:
//   - Returns "mock failure" if the mock builder configured an intentional exception.
//
// Side Effects:
//   - Mutates the internal call history array to record the execution timeline for test assertions.
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
