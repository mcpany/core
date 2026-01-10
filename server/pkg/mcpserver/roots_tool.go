// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"fmt"

	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/types/known/structpb"
)

// RootsTool implements the Tool interface for listing roots.
type RootsTool struct {
	tool    *v1.Tool
	mcpTool *mcp.Tool
}

// NewRootsTool creates a new RootsTool.
//
// Returns:
//   *RootsTool: A new instance of RootsTool.
func NewRootsTool() *RootsTool {
	name := "mcp:list_roots"
	displayName := "List Roots"
	description := "Lists the roots available on the client side."
	serviceID := "builtin"
	t := &v1.Tool{
		Name:        &name,
		DisplayName: &displayName,
		Description: &description,
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"type": structpb.NewStringValue("object"),
			},
		},
		ServiceId: &serviceID,
	}
	mcpTool, _ := tool.ConvertProtoToMCPTool(t)
	return &RootsTool{
		tool:    t,
		mcpTool: mcpTool,
	}
}

// Tool returns the protobuf definition of the tool.
//
// Returns:
//   *v1.Tool: The protobuf tool definition.
func (t *RootsTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP tool definition.
//
// Returns:
//   *mcp.Tool: The MCP tool definition.
func (t *RootsTool) MCPTool() *mcp.Tool {
	return t.mcpTool
}

// Execute executes the tool.
//
// Parameters:
//   ctx: The context for the execution.
//   _: The execution request (unused).
//
// Returns:
//   any: The result of the execution (list of roots).
//   error: An error if the execution fails.
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
//   *configv1.CacheConfig: The cache configuration (nil for this tool).
func (t *RootsTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

// Verify that RootsTool implements tool.Tool.
var _ tool.Tool = (*RootsTool)(nil)
