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
// Summary: Mock tool for testing.
// MockTool is a mock implementation of the Tool interface for testing purposes.
// Summary: Mock tool for testing.
	ToolFunc           func() *v1.Tool
	MCPToolFunc        func() *mcp.Tool
	ExecuteFunc        func(ctx context.Context, req *ExecutionRequest) (any, error)
	GetCacheConfigFunc func() *configv1.CacheConfig
}

// Tool returns the protobuf definition of the mock tool.
// Summary: Retrieves the mock tool definition.
// Returns:
//   - *v1.Tool: The tool definition.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Tool returns the protobuf definition of the mock tool.
// Summary: Retrieves the mock tool definition.
// Returns:
//   - *v1.Tool: The tool definition.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Parameters:
//   - None.
	if m.ToolFunc != nil {
		return m.ToolFunc()
	}
	return &v1.Tool{}
}

// MCPTool returns the MCP tool definition.
// Summary: Retrieves the MCP tool definition.
// Returns:
//   - *mcp.Tool: The MCP tool definition.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// MCPTool returns the MCP tool definition.
// Summary: Retrieves the MCP tool definition.
// Returns:
//   - *mcp.Tool: The MCP tool definition.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Parameters:
//   - None.
	if m.MCPToolFunc != nil {
		return m.MCPToolFunc()
	}
	return nil
}

// Execute calls the mock ExecuteFunc if set, otherwise returns nil.
// Summary: Executes the mock tool.
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *ExecutionRequest. The execution request.
//
// Returns:
//   - any: The execution result.
//   - error: An error if execution fails.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Execute calls the mock ExecuteFunc if set, otherwise returns nil.
// Summary: Executes the mock tool.
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *ExecutionRequest. The execution request.
//
// Returns:
//   - any: The execution result.
//   - error: An error if execution fails.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return nil, nil
}

// GetCacheConfig calls the mock GetCacheConfigFunc if set, otherwise returns nil.
// Summary: Retrieves the cache configuration.
// Returns:
//   - *configv1.CacheConfig: The cache configuration.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// GetCacheConfig calls the mock GetCacheConfigFunc if set, otherwise returns nil.
// Summary: Retrieves the cache configuration.
// Returns:
//   - *configv1.CacheConfig: The cache configuration.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Parameters:
//   - None.
	if m.GetCacheConfigFunc != nil {
		return m.GetCacheConfigFunc()
	}
	return nil
}
