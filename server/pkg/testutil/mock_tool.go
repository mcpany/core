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

// MockTool represents the public MockTool entity.
//
// Summary: Provides tool execution capabilities for tool.
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
	ExecuteFunc func(ctx context.Context, req *tool.ExecutionRequest) (any, error)
}

// Tool serves as a public interface for interacting with Tool.
//
// Summary: Tool the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *MockTool) Tool() *v1.Tool {
	return v1.Tool_builder{
		Name: proto.String("mock-tool"),
	}.Build()
}

// Execute serves as a public interface for interacting with Execute.
//
// Summary: Execute the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// GetCacheConfig serves as a public interface for interacting with GetCacheConfig.
//
// Summary: Fetches and returns the underlying cache config from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
