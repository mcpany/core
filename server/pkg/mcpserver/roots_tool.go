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
// It provides a built-in tool ("mcp:list_roots") that allows the server to query the client
// for available filesystem roots.
//
// Summary: Built-in tool for listing client filesystem roots.
type RootsTool struct {
	tool    *v1.Tool
	mcpTool *mcp.Tool
}

// NewRootsTool creates a new instance of the RootsTool.
//
// Summary: Initializes NewRootsTool operation.
//
// Returns:
//   - *RootsTool: A new instance of RootsTool.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
// Summary: Executes Tool operation.
//
// Returns:
//   - *v1.Tool: The protobuf tool definition.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *RootsTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP-compliant tool definition.
//
// Summary: Executes MCPTool operation.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *RootsTool) MCPTool() *mcp.Tool {
	return t.mcpTool
}

// Execute executes the "mcp:list_roots" tool.
//
// Summary: Executes Execute operation.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - req: *tool.ExecutionRequest. The execution request.
//
// Returns:
//   - any: The result of the roots list operation.
//   - error: An error if the list operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
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

// GetCacheConfig returns the caching configuration for this tool.
//
// Summary: Retrieves GetCacheConfig operation.
//
// Returns:
//   - *configv1.CacheConfig: The caching configuration.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *RootsTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

// Verify that RootsTool implements tool.Tool.
var _ tool.Tool = (*RootsTool)(nil)
