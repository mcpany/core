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

	// Fallback mock data if ExecuteFunc is not set.
	// Returns a complex, nested JSON array structure for testing table rendering.
	return []map[string]any{
		{
			"id":   1,
			"name": "Alice Smith",
			"contact": map[string]any{
				"email": "alice@example.com",
				"phone": map[string]any{
					"mobile": "555-1234",
					"home":   "555-5678",
				},
			},
			"tags": []string{"admin", "active"},
			"metadata": map[string]any{
				"lastLogin": "2023-01-01T12:00:00Z",
				"preferences": map[string]any{
					"theme": "dark",
					"notifications": true,
				},
			},
		},
		{
			"id":   2,
			"name": "Bob Jones",
			"contact": map[string]any{
				"email": "bob@example.com",
			},
			"tags": []string{"user"},
			"metadata": map[string]any{
				"lastLogin": "2023-01-02T12:00:00Z",
				"preferences": map[string]any{
					"theme": "light",
					"notifications": false,
				},
			},
		},
		{
			"id":   3,
			"name": "Charlie Brown",
			"contact": nil,
			"tags": []string{},
			"metadata": map[string]any{
				"lastLogin": "2023-01-03T12:00:00Z",
			},
		},
	}, nil
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
