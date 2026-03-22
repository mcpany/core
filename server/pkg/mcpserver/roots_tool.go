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
//   - []Root: A list of available filesystem roots.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns "no active session available" if the session is missing.
//
// Side Effects:
//   - None.

// IsStreaming indicates whether the tool supports streaming.
//
// Summary: Returns false as the Roots tool does not support streaming.
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

// StreamExecute is an unsupported streaming execution method for the Roots tool.
//
// Summary: Returns nil as streaming is not supported.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (*tool.ExecutionRequest): The execution request parameters.
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

// Execute retrieves the list of filesystem roots from the client.
//
// Summary: Fetches filesystem roots using the active MCP session.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - _ (*tool.ExecutionRequest): The execution request parameters (unused).
//
// Returns:
//   - any: The parsed list of Roots from the client.
//   - error: An error if the session is inactive or if the client returns an error.
//
// Errors:
//   - Returns an error if no active session is available.
//   - Returns an error if the roots cannot be listed.
//   - Returns an error if the client responds with an error.
//
// Side Effects:
//   - Issues an RPC call to the client to request root information.
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
