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
type MockTool struct {
	ExecuteFunc func(ctx context.Context, req *tool.ExecutionRequest) (any, error)
}

// Tool returns a basic tool definition for the mock tool.
//
// Summary: Returns the tool definition.
//
// Returns:
//   - (*v1.Tool): The tool definition protobuf.
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
//   - ctx (context.Context): The request context.
//   - req (*tool.ExecutionRequest): The execution request.
//
// Returns:
//   - (any): The result of the execution.
//   - (error): An error if the execution fails.
func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// GetCacheConfig returns nil for the mock tool.
//
// Summary: Returns the cache configuration.
//
// Returns:
//   - (*configv1.CacheConfig): Always nil for this mock.
func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
