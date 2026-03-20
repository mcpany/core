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
// Summary: RootsTool implements the Tool interface for listing roots.
type RootsTool struct {
	tool    *v1.Tool
	mcpTool *mcp.Tool
}

// NewRootsTool creates a new instance of the RootsTool.
//
// Summary: NewRootsTool creates a new instance of the RootsTool.
//
// Parameters:
//   - None.
//
// Returns:
//   - *RootsTool: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
// Summary: Tool returns the protobuf definition of the tool.
//
// Parameters:
//   - None.
//
// Returns:
//   - *v1.Tool: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (t *RootsTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP-compliant tool definition.
//
// Summary: MCPTool returns the MCP-compliant tool definition.
//
// Parameters:
//   - None.
//
// Returns:
//   - *mcp.Tool: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (t *RootsTool) MCPTool() *mcp.Tool {
	return t.mcpTool
}

// Execute executes the "mcp:list_roots" tool.
//
// Summary: Execute executes the "mcp:list_roots" tool.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - _ (*tool.ExecutionRequest): The provided _ data.
//
// Returns:
//   - any: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
// Summary: GetCacheConfig returns the caching configuration for this tool.
//
// Parameters:
//   - None.
//
// Returns:
//   - *configv1.CacheConfig: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (t *RootsTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

// Verify that RootsTool implements tool.Tool.
var _ tool.Tool = (*RootsTool)(nil)
