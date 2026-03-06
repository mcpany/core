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

// MockTool is a mock implementation of the tool.Tool interface for testing.
//
// Summary: Mock tool for unit testing.
type MockTool struct {
	ExecuteFunc func(ctx context.Context, req *tool.ExecutionRequest) (any, error)
}

// Tool returns a basic tool definition for the mock tool.
//
// Summary: Returns the tool definition.
//
// Returns:
//   - *v1.Tool: A minimal tool definition.
func (m *MockTool) Tool() *v1.Tool {
	return v1.Tool_builder{
		Name: proto.String("mock-tool"),
	}.Build()
}

// Execute calls the mock ExecuteFunc if set, otherwise returns nil.
//
// Summary: Executes the mock tool logic.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *tool.ExecutionRequest. The tool execution request.
//
// Returns:
//   - any: The result from ExecuteFunc.
//   - error: The error from ExecuteFunc.
//
// Side Effects:
//   - Invokes the injected ExecuteFunc.
func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// GetCacheConfig returns nil for the mock tool.
//
// Summary: Returns cache configuration (nil for mock).
//
// Returns:
//   - *configv1.CacheConfig: Always nil.
func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
