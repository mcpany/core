// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package tool defines the interface for tools that can be executed by the upstream service.
package tool

import (
	"context"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/types/known/structpb"
)

type baseTool struct {
	tool          *v1.Tool
	mcpTool       *mcp.Tool
	mcpToolOnce   sync.Once
	serviceConfig *configv1.UpstreamServiceConfig
	callable      Callable
}

func newBaseTool(toolDef *configv1.ToolDefinition, serviceConfig *configv1.UpstreamServiceConfig, callable Callable, inputSchema, outputSchema *structpb.Struct) (*baseTool, error) {
	pbTool, err := ConvertToolDefinitionToProto(toolDef, inputSchema, outputSchema)
	if err != nil {
		return nil, err
	}
	return &baseTool{
		tool:          pbTool,
		serviceConfig: serviceConfig,
		callable:      callable,
	}, nil
}

// Tool serves as a public interface for interacting with Tool.
//
// Summary: Tool the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (t *baseTool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool serves as a public interface for interacting with MCPTool.
//
// Summary: Mcp the tool appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (t *baseTool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
		var err error
		t.mcpTool, err = ConvertProtoToMCPTool(t.tool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert tool to MCP tool", "toolName", t.tool.GetName(), "error", err)
		}
	})
	return t.mcpTool
}

// IsStreaming serves as a public interface for interacting with IsStreaming.
//
// Summary: Checks condition indicating whether the target is streaming.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (t *baseTool) IsStreaming() bool {
	return false
}

// StreamExecute serves as a public interface for interacting with StreamExecute.
//
// Summary: Stream the execute appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (t *baseTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
	return nil, nil // Should be implemented by embedding struct if supported
}

// GetCacheConfig serves as a public interface for interacting with GetCacheConfig.
//
// Summary: Fetches and returns the underlying cache config from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (t *baseTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
