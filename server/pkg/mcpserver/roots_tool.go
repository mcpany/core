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

// NewRootsTool creates a new instance of the RootsTool.
//
// Returns:
//   - *RootsTool: A new instance of RootsTool.
//
// Side Effects:
//   - None.
//
// Summary: Initializes NewRootsTool operation.
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
//   - *v1.Tool: The protobuf tool definition.
//
// Side Effects:
//   - None.
//
// Summary: Executes Tool operation.
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
func (t *RootsTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP-compliant tool definition.
//
// Returns:
//   - *mcp.Tool: The MCP tool definition.
//
// Side Effects:
//   - None.
//
// Summary: Executes MCPTool operation.
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

// IsStreaming indicates whether the RootsTool supports streaming.
//
// Summary: Returns false as RootsTool does not support streaming.
//
// Parameters:
//   - None.
//
// Returns:
//   - bool: Always returns false.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *RootsTool) IsStreaming() bool {
	return false
}

// StreamExecute is not implemented for RootsTool.
//
// Summary: Returns nil channels as streaming is not supported.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - req (*tool.ExecutionRequest): The execution request.
//
// Returns:
//   - <-chan any: Always nil.
//   - error: Always nil.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *RootsTool) StreamExecute(ctx context.Context, req *tool.ExecutionRequest) (<-chan any, error) {
	return nil, nil
}

// Execute triggers the RootsTool to retrieve roots from the active session.
//
// Summary: Retrieves roots from the active MCP session.
//
// Parameters:
//   - ctx (context.Context): The execution context containing the active session.
//   - _ (*tool.ExecutionRequest): Unused execution request.
//
// Returns:
//   - any: The roots result from the session.
//   - error: An error if no active session is found or the list operation fails.
//
// Errors:
//   - Returns an error if no active session is available.
//   - Returns an error if listing roots fails.
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
// Returns:
//   - *configv1.CacheConfig: Always nil (caching disabled).
//
// Side Effects:
//   - None.
//
// Summary: Retrieves GetCacheConfig operation.
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
func (t *RootsTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

// Verify that RootsTool implements tool.Tool.
var _ tool.Tool = (*RootsTool)(nil)
