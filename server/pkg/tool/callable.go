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
//
// Summary: Represents a CallableTool.
type CallableTool struct {
	*baseTool
}

// NewCallableTool provides newcallabletool functionality.
//
// Summary: NewCallableTool.
//
// Parameters.
//   - toolDef: The parameter.
//   - serviceConfig: The parameter.
//   - callable: The parameter.
//   - inputSchema: The parameter.
//   - outputSchema: The parameter.
//
// Returns.
//   - result: The result.
//
// Parameters:
//   - toolDef: *configv1.ToolDefinition.
//   - serviceConfig: *configv1.UpstreamServiceConfig.
//   - callable: Callable.
//   - inputSchema: unknown.
//   - outputSchema: *structpb.Struct.
//
// Returns:
//   - *CallableTool.
//   - error.
func NewCallableTool(toolDef *configv1.ToolDefinition, serviceConfig *configv1.UpstreamServiceConfig, callable Callable, inputSchema, outputSchema *structpb.Struct) (*CallableTool, error) {
	base, err := newBaseTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
	if err != nil {
		return nil, err
	}
	return &CallableTool{base}, nil
}

// Execute provides execute functionality.
//
// Summary: Execute.
//
// Parameters.
//   - ctx: The parameter.
//   - req: The parameter.
//
// Returns.
//   - result: The result.
//
// Parameters:
//   - ctx: context.Context.
//   - req: *ExecutionRequest.
//
// Returns:
//   - any.
//   - error.
func (t *CallableTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	return t.callable.Call(ctx, req)
}

// Callable provides callable functionality.
//
// Summary: Callable.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
//
// Parameters:
//   - None.
//
// Returns:
//   - Callable.
func (t *CallableTool) Callable() Callable {
	return t.callable
}
