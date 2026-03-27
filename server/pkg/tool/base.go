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
// Summary: Retrieves the protobuf definition.
//
// Returns:
//   - *v1.Tool: The protobuf tool definition.
func (t *baseTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP tool definition.
//
// Summary: Retrieves the MCP-compliant tool definition.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
//
// Side Effects:
//   - Lazily converts the proto definition to MCP format on first call.
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

// GetCacheConfig returns the cache configuration for the tool, or nil if caching is disabled.
//
// Summary: Retrieves the cache configuration (always nil for baseTool).
//
// Returns:
//   - *configv1.CacheConfig: Always returns nil.
// IsStreaming returns true if the tool supports streaming execution.
//
// Summary: Checks if the tool supports streaming execution.
//
// Returns:
//   - bool: True if streaming is supported.
func (t *baseTool) IsStreaming() bool {
	return false
}

// StreamExecute handles the streaming execution of the tool.
//
// Summary: Executes the tool in streaming mode.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request payload.
//
// Returns:
//   - <-chan any: A channel that emits streaming results.
//   - error: An error if the operation fails or streaming is not supported.
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
