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

// Tool provides tool functionality.
//
// Summary: Tool.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (t *baseTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool provides mcptool functionality.
//
// Summary: MCPTool.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
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

// GetCacheConfig provides getcacheconfig functionality.
//
// Summary: GetCacheConfig.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (t *baseTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
