// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// CallableTool implements the Tool interface for a tool that is executed by a.
//
// Summary: implements the Tool interface for a tool that is executed by a.
type CallableTool struct {
	*baseTool
}

// NewCallableTool creates a new CallableTool.
//
// Summary: creates a new CallableTool.
//
// Parameters:
//   - toolDef: *configv1.ToolDefinition. The toolDef.
//   - serviceConfig: *configv1.UpstreamServiceConfig. The serviceConfig.
//   - callable: Callable. The callable.
//   - inputSchema: *structpb.Struct. The inputSchema.
//   - outputSchema: *structpb.Struct. The outputSchema.
//
// Returns:
//   - *CallableTool: The *CallableTool.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func NewCallableTool(toolDef *configv1.ToolDefinition, serviceConfig *configv1.UpstreamServiceConfig, callable Callable, inputSchema, outputSchema *structpb.Struct) (*CallableTool, error) {
	base, err := newBaseTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
	if err != nil {
		return nil, err
	}
	return &CallableTool{base}, nil
}

// Execute handles the execution of the tool.
//
// Summary: handles the execution of the tool.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - req: *ExecutionRequest. The req.
//
// Returns:
//   - any: The any.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (t *CallableTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	return t.callable.Call(ctx, req)
}

// Callable returns the underlying Callable of the tool.
//
// Summary: returns the underlying Callable of the tool.
//
// Parameters:
//   None.
//
// Returns:
//   - Callable: The Callable.
func (t *CallableTool) Callable() Callable {
	return t.callable
}
