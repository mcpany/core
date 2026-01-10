// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SimpleTool implements the Tool interface for testing and simple use cases.
type SimpleTool struct {
	tool     *v1.Tool
	callback func(ctx context.Context, args json.RawMessage) (*Result, error)
}

// Result represents a simplified result for SimpleTool.
type Result struct {
	Content []any
}

// NewSimpleTool creates a new SimpleTool.
func NewSimpleTool(tool *v1.Tool, callback func(ctx context.Context, args json.RawMessage) (*Result, error)) *SimpleTool {
	return &SimpleTool{
		tool:     tool,
		callback: callback,
	}
}

// Tool returns the protobuf definition of the tool.
func (t *SimpleTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP tool definition.
func (t *SimpleTool) MCPTool() *mcp.Tool {
	mcpTool, _ := ConvertProtoToMCPTool(t.tool)
	return mcpTool
}

// Execute runs the tool with the provided context and request.
func (t *SimpleTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	if t.callback != nil {
		return t.callback(ctx, req.ToolInputs)
	}
	return map[string]any{"status": "ok"}, nil
}

// GetCacheConfig returns nil for SimpleTool.
func (t *SimpleTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
