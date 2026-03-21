// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package tool defines the interface for tools that can be executed by the upstream service.
package tool

import (
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

// Summary: Tool returns the protobuf definition of the tool. Retrieves the protobuf definition.
//
// Parameters:
//   - None.
//
// Returns:
//   - *v1.Tool: The resulting *v1.Tool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *baseTool) Tool() *v1.Tool {
	return t.tool
}

// Summary: MCPTool returns the MCP tool definition. Retrieves the MCP-compliant tool definition.
//
// Parameters:
//   - None.
//
// Returns:
//   - *mcp.Tool: The resulting *mcp.Tool.
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

// Summary: GetCacheConfig returns the cache configuration for the tool, or nil if caching is disabled. Retrieves the cache configuration (always nil for baseTool).
//
// Parameters:
//   - None.
//
// Returns:
//   - *configv1.CacheConfig: The resulting *configv1.CacheConfig.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *baseTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
