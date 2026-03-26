// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// CallableTool callableTool represents a callable tool.
//
// Summary: CallableTool represents a callable tool.
type CallableTool struct {
	*baseTool
}

// NewCallableTool creates a new CallableTool.
//
// Summary: Creates a new tool that wraps a Callable interface.
//
// Parameters: - None.
//   - toolDef: *configv1.ToolDefinition. The definition of the tool.
//   - serviceConfig: *configv1.UpstreamServiceConfig. The configuration of the service the tool belongs to.
//   - callable: Callable. The callable implementation for execution.
//   - inputSchema: *structpb.Struct. The input schema for the tool.
//   - outputSchema: *structpb.Struct. The output schema for the tool.
//
// Returns: - None.
//   - *CallableTool: A pointer to the created CallableTool.
//   - error: An error if creation fails.
func NewCallableTool(toolDef *configv1.ToolDefinition, serviceConfig *configv1.UpstreamServiceConfig, callable Callable, inputSchema, outputSchema *structpb.Struct) (*CallableTool, error) {
	base, err := newBaseTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
	if err != nil {
		return nil, err
	}
	return &CallableTool{base}, nil
}

// Execute handles the execution of the tool.
//
// Summary: Executes the underlying callable.
//
// Parameters: - None.
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request object containing parameters.
//
// Returns: - None.
//   - any: The result of the execution.
//   - error: An error if the operation fails.
func (t *CallableTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	return t.callable.Call(ctx, req)
}

// Callable callable callable.
//
// Summary: Callable callable.
//
// Parameters:
//   - None.
//
// Returns:
//   - Callable: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *CallableTool) Callable() Callable {
	return t.callable
}
