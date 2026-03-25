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
// Parameters:
//   None.
//
// Returns:
//   - *RootsTool: The result.
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

// Tool tool tool.
//
// Summary: Tool tool.
//
// Parameters:
//   None.
//
// Returns:
//   - *v1.Tool: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *RootsTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool mCPTool mcp tool.
//
// Summary: MCPTool mcp tool.
//
// Parameters:
//   None.
//
// Returns:
//   - *mcp.Tool: The result.
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
// It retrieves the current MCP session from the context and requests the client
// to list its roots.
//
// Parameters:
//   - ctx (context.Context): The request context, must contain an active MCP session.
//   - _ (*tool.ExecutionRequest): The execution request parameters (unused as this tool takes no inputs).
//
// Returns:
//   - any: The result of the roots list operation (typically a list of roots).
//   - error: An error if the session is missing or the list operation fails.
//
// Side Effects:
//   - Sends a "roots/list" request to the client.
//
// Summary: Executes Execute operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.

// IsStreaming indicates whether this tool supports streaming execution.
//
// Summary: Checks if the RootsTool supports streaming.
//
// Parameters:
//   - None.
//
// Returns:
//   - bool: Always false for this tool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *RootsTool) IsStreaming() bool {
	return false
}

// StreamExecute executes the tool in a streaming context.
//
// Summary: Executes the tool in a streaming context.
//
// Parameters:
//   - ctx (context.Context): The context for execution.
//   - req (*tool.ExecutionRequest): The execution request parameters.
//
// Returns:
//   - <-chan any: Always nil for this tool.
//   - error: Always nil for this tool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *RootsTool) StreamExecute(ctx context.Context, req *tool.ExecutionRequest) (<-chan any, error) {
	return nil, nil
}

// Execute executes the tool.
//
// Summary: Executes the "mcp:list_roots" tool.
//
// Parameters:
//   - ctx (context.Context): The context for execution, containing an active MCP session.
//   - _ (*tool.ExecutionRequest): The execution request parameters (unused).
//
// Returns:
//   - any: The result of listing roots.
//   - error: An error if the session is missing or the list operation fails.
//
// Errors:
//   - Returns error if the active session is missing.
//   - Returns error if listing roots fails.
//
// Side Effects:
//   - Sends a "roots/list" request to the client.
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
// Parameters:
//   None.
//
// Returns:
//   - *configv1.CacheConfig: The result.
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
