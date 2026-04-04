// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package tool defines the interface for tools that can be executed by the upstream service.
package tool

import (
	"context"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/types/known/structpb"
)

type baseTool struct {
	tool          *v1.Tool
	mcpTool       *mcp.Tool
	mcpToolOnce   sync.Once
	serviceConfig *configv1.UpstreamServiceConfig
	callable      Callable
}

func newBaseTool(toolDef *configv1.ToolDefinition, serviceConfig *configv1.UpstreamServiceConfig, callable Callable, inputSchema, outputSchema *structpb.Struct) (*baseTool, error) {
	pbTool, err := ConvertToolDefinitionToProto(toolDef, inputSchema, outputSchema)
	if err != nil {
		return nil, err
	}
	return &baseTool{
		tool:          pbTool,
		serviceConfig: serviceConfig,
		callable:      callable,
	}, nil
}

// Tool returns the protobuf definition of the tool.
//
// Summary: Returns the protobuf definition of the tool.
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

func (t *baseTool) Tool() *v1.Tool {
	return t.tool
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

func (t *baseTool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
		var err error
		t.mcpTool, err = ConvertProtoToMCPTool(t.tool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert tool to MCP tool", "toolName", t.tool.GetName(), "error", err)
		}
	})
	return t.mcpTool
}

// IsStreaming getCacheConfig returns the cache configuration for the tool, or nil if caching is disabled.
//
// Summary: GetCacheConfig returns the cache configuration for the tool, or nil if caching is disabled.
//
// Parameters:
//   - None.
//
// Returns:
//   - bool: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (t *baseTool) IsStreaming() bool {
	return false
}

// StreamExecute handles the streaming execution of the tool.
//
// Summary: Handles the streaming execution of the tool.
//
// Parameters:
//   - ctx (context.Context): Parameter.
//   - req (*ExecutionRequest): Parameter.
//
// Returns:
//   - <-chan any: Return value.
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

func (t *baseTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
	return nil, nil // Should be implemented by embedding struct if supported
}

// GetCacheConfig returns the cache configuration for the tool.
//
// Summary: Retrieves the cache configuration for the tool.
//
// Parameters:
//   - None.
//
// Returns:
//   - *configv1.CacheConfig: Always returns nil for baseTool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *baseTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
