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

// Summary: RootsTool implements the Tool interface for listing roots. It provides a built-in tool ("mcp:list_roots") that allows the server to query the client for available filesystem roots. Represents a RootsTool.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type RootsTool struct {
	tool    *v1.Tool
	mcpTool *mcp.Tool
}

// Summary: NewRootsTool creates a new instance of the RootsTool.
//
// Parameters:
//   - None.
//
// Returns:
//   - *RootsTool: The resulting *RootsTool.
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

// Summary: Tool returns the protobuf definition of the tool.
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
func (t *RootsTool) Tool() *v1.Tool {
	return t.tool
}

// Summary: MCPTool returns the MCP-compliant tool definition.
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
func (t *RootsTool) MCPTool() *mcp.Tool {
	return t.mcpTool
}

// Summary: Execute executes the "mcp:list_roots" tool. It retrieves the current MCP session from the context and requests the client to list its roots.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - _ (*tool.ExecutionRequest): The _ parameter.
//
// Returns:
//   - any: The resulting any.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
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

// Summary: GetCacheConfig returns the caching configuration for this tool.
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
func (t *RootsTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

// Verify that RootsTool implements tool.Tool.
var _ tool.Tool = (*RootsTool)(nil)
