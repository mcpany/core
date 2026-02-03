// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// RootsTool implements the Tool interface for listing roots.
//
// It provides a "mcp:list_roots" tool that allows the server or other tools to
// query the roots available on the client side.
type RootsTool struct {
	tool    *v1.Tool
	mcpTool *mcp.Tool
}

// NewRootsTool creates a new RootsTool.
//
// Returns:
//   - *RootsTool: A new instance of RootsTool.
func NewRootsTool() *RootsTool {
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": structpb.NewStringValue("object"),
		},
	}
	t := v1.Tool_builder{
		Name:        proto.String("mcp:list_roots"),
		DisplayName: proto.String("List Roots"),
		Description: proto.String("Lists the roots available on the client side."),
		InputSchema: inputSchema,
		ServiceId:   proto.String("builtin"),
	}.Build()

	mcpTool, _ := tool.ConvertProtoToMCPTool(t)
	return &RootsTool{
		tool:    t,
		mcpTool: mcpTool,
	}
}

// Tool returns the protobuf definition of the tool.
//
// Returns:
//   - *v1.Tool: The protobuf definition.
func (t *RootsTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP tool definition.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
func (t *RootsTool) MCPTool() *mcp.Tool {
	return t.mcpTool
}

// Execute executes the tool.
//
// It calls the client's "roots/list" method to retrieve the roots.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *tool.ExecutionRequest. The execution request (unused inputs).
//
// Returns:
//   - any: The result of the "roots/list" call.
//   - error: An error if the session is missing or the call fails.
func (t *RootsTool) Execute(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {
	session, ok := tool.GetSession(ctx)
	if !ok {
		return nil, fmt.Errorf("no active session available")
	}

	rootsResult, err := session.ListRoots(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list roots: %w", err)
	}

	return rootsResult, nil
}

// GetCacheConfig returns the cache configuration for the tool.
//
// Returns:
//   - *configv1.CacheConfig: Always returns nil as this tool shouldn't be cached aggressively.
func (t *RootsTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

// Verify that RootsTool implements tool.Tool.
var _ tool.Tool = (*RootsTool)(nil)
