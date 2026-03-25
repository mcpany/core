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
// Summary: Represents a RootsTool.
type RootsTool struct {
	tool    *v1.Tool
	mcpTool *mcp.Tool
}

// NewRootsTool creates a new roots tool.
//
// Summary: Creates a new roots tool.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *RootsTool: The result.
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

// Tool tool tool.
//
// Summary: Tool tool.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *v1.Tool: The result.
func (t *RootsTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool mCPTool mcp tool.
//
// Summary: MCPTool mcp tool.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *mcp.Tool: The result.
func (t *RootsTool) MCPTool() *mcp.Tool {
	return t.mcpTool
}

// Execute executes the operation.
//
// Summary: Executes the operation.
//
// Parameters: - None.
//   - ctx (context.Context): The context for the request.
//   - _ (*tool.ExecutionRequest): Unused parameter.
//
// Returns: - None.
//   - any: The result.
//   - error: An error if the operation fails.
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

// GetCacheConfig retrieves the cache config.
//
// Summary: Retrieves the cache config.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *configv1.CacheConfig: The result.
func (t *RootsTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

// Verify that RootsTool implements tool.Tool.
var _ tool.Tool = (*RootsTool)(nil)
