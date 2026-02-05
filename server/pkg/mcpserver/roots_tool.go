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
// Summary: Provides a built-in tool for listing the client's root directories (filesystem roots).
type RootsTool struct {
	tool    *v1.Tool
	mcpTool *mcp.Tool
}

// NewRootsTool creates a new RootsTool.
//
// Summary: Initializes a new RootsTool instance.
//
// Returns:
//   - *RootsTool: A pointer to the new RootsTool.
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
//   - *mcp.Tool: The MCP SDK tool definition.
func (t *RootsTool) MCPTool() *mcp.Tool {
	return t.mcpTool
}

// Execute executes the tool.
//
// Summary: Executes the "mcp:list_roots" tool to fetch roots from the client.
//
// Parameters:
//   - ctx: context.Context. The context for the execution.
//   - _ : *tool.ExecutionRequest. The execution request (unused, as this tool takes no arguments).
//
// Returns:
//   - any: The result of the root listing (usually *mcp.ListRootsResult).
//   - error: An error if the operation fails or no session is available.
//
// Throws/Errors:
//   - Returns an error if no active MCP session is found in the context.
//   - Returns an error if the client fails to return the list of roots.
//
// Side Effects:
//   - Sends a "roots/list" request to the connected client.
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

// GetCacheConfig returns nil as this tool shouldn't be cached aggressively or depends on client state.
//
// Summary: Returns the cache configuration for this tool.
//
// Returns:
//   - *configv1.CacheConfig: Always nil, as this tool depends on dynamic client state.
func (t *RootsTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

// Verify that RootsTool implements tool.Tool.
var _ tool.Tool = (*RootsTool)(nil)
