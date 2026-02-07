// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// CallableTool implements the Tool interface for a tool that is executed by a
// Callable.
type CallableTool struct {
	*baseTool
}

// NewCallableTool creates a new CallableTool.
//
// Summary: Initializes a new CallableTool instance.
//
// Parameters:
//   - toolDef: *configv1.ToolDefinition. The definition of the tool.
//   - serviceConfig: *configv1.UpstreamServiceConfig. The configuration of the service the tool belongs to.
//   - callable: Callable. The callable implementation for execution.
//   - inputSchema: *structpb.Struct. The input schema for the tool.
//   - outputSchema: *structpb.Struct. The output schema for the tool.
//
// Returns:
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
// Summary: Executes the tool using the underlying callable.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request object containing inputs.
//
// Returns:
//   - any: The result of the execution.
//   - error: An error if the operation fails.
func (t *CallableTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	return t.callable.Call(ctx, req)
}

// Callable returns the underlying Callable of the tool.
//
// Summary: Retrieves the underlying Callable instance.
//
// Returns:
//   - Callable: The callable instance.
func (t *CallableTool) Callable() Callable {
	return t.callable
}
