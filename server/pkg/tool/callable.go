// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// Summary: CallableTool implements the Tool interface for a tool that is executed by a Callable. Represents a CallableTool.
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
type CallableTool struct {
	*baseTool
}

// Summary: NewCallableTool creates a new CallableTool. Creates a new tool that wraps a Callable interface.
//
// Parameters:
//   - toolDef (*configv1.ToolDefinition): The toolDef parameter.
//   - serviceConfig (*configv1.UpstreamServiceConfig): The serviceConfig parameter.
//   - callable (Callable): The callable parameter.
//   - inputSchema (*structpb.Struct): The inputSchema parameter.
//   - outputSchema (*structpb.Struct): The outputSchema parameter.
//
// Returns:
//   - *CallableTool: The resulting *CallableTool.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func NewCallableTool(toolDef *configv1.ToolDefinition, serviceConfig *configv1.UpstreamServiceConfig, callable Callable, inputSchema, outputSchema *structpb.Struct) (*CallableTool, error) {
	base, err := newBaseTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
	if err != nil {
		return nil, err
	}
	return &CallableTool{base}, nil
}

// Summary: Execute handles the execution of the tool. Executes the underlying callable.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - req (*ExecutionRequest): The req parameter.
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
func (t *CallableTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	return t.callable.Call(ctx, req)
}

// Summary: Callable returns the underlying Callable of the tool. Retrieves the underlying Callable interface.
//
// Parameters:
//   - None.
//
// Returns:
//   - Callable: The resulting Callable.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (t *CallableTool) Callable() Callable {
	return t.callable
}
