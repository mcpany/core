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

// RootsTool implements the Tool interface for listing client-side roots.
//
// Summary: A built-in tool that allows the server to query the client for its file system roots.
type RootsTool struct {
	tool    *v1.Tool
	mcpTool *mcp.Tool
}

// NewRootsTool creates a new RootsTool.
//
// Summary: Initializes the "mcp:list_roots" tool.
//
// Returns:
//   - *RootsTool: The initialized tool instance.
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
// Summary: Retrieves the protobuf representation of the tool.
//
// Returns:
//   - *v1.Tool: The protobuf tool definition.
func (t *RootsTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP tool definition.
//
// Summary: Retrieves the MCP SDK representation of the tool.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
func (t *RootsTool) MCPTool() *mcp.Tool {
	return t.mcpTool
}

// Execute executes the tool.
//
// Summary: Invokes the "list_roots" capability on the client session.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - _: *tool.ExecutionRequest. Unused parameters as this tool takes no arguments.
//
// Returns:
//   - any: The result of the client's list_roots call (*mcp.ListRootsResult).
//   - error: An error if no session is available or the call fails.
//
// Throws/Errors:
//   - Returns an error if "no active session available".
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

// GetCacheConfig returns the cache configuration for this tool.
//
// Summary: Returns nil to indicate this tool should not be cached.
//
// Returns:
//   - *configv1.CacheConfig: Always nil.
func (t *RootsTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

// Verify that RootsTool implements tool.Tool.
var _ tool.Tool = (*RootsTool)(nil)
