// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// CallableTool - Auto-generated documentation.
//
// Summary: CallableTool implements the Tool interface for a tool that is executed by a
//
// Fields:
//   - Various fields for CallableTool.
type CallableTool struct {
	*baseTool
}

// NewCallableTool creates a new CallableTool. Summary: Creates a new tool that wraps a Callable interface. Parameters: - toolDef: *configv1.ToolDefinition. The definition of the tool. - serviceConfig: *configv1.UpstreamServiceConfig. The configuration of the service the tool belongs to. - callable: Callable. The callable implementation for execution. - inputSchema: *structpb.Struct. The input schema for the tool. - outputSchema: *structpb.Struct. The output schema for the tool. Returns: - *CallableTool: A pointer to the created CallableTool. - error: An error if creation fails.
//
// Summary: NewCallableTool creates a new CallableTool. Summary: Creates a new tool that wraps a Callable interface. Parameters: - toolDef: *configv1.ToolDefinition. The definition of the tool. - serviceConfig: *configv1.UpstreamServiceConfig. The configuration of the service the tool belongs to. - callable: Callable. The callable implementation for execution. - inputSchema: *structpb.Struct. The input schema for the tool. - outputSchema: *structpb.Struct. The output schema for the tool. Returns: - *CallableTool: A pointer to the created CallableTool. - error: An error if creation fails.
//
// Parameters:
//   - toolDef (*configv1.ToolDefinition): The tool def parameter used in the operation.
//   - serviceConfig (*configv1.UpstreamServiceConfig): The configuration settings to be applied.
//   - callable (Callable): The callable parameter used in the operation.
//   - _ (inputSchema): An unnamed parameter of type inputSchema.
//   - outputSchema (*structpb.Struct): The output schema parameter used in the operation.
//
// Returns:
//   - (*CallableTool): The resulting CallableTool object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
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

// Execute handles the execution of the tool. Summary: Executes the underlying callable. Parameters: - ctx: context.Context. The context for the request. - req: *ExecutionRequest. The request object containing parameters. Returns: - any: The result of the execution. - error: An error if the operation fails.
//
// Summary: Execute handles the execution of the tool. Summary: Executes the underlying callable. Parameters: - ctx: context.Context. The context for the request. - req: *ExecutionRequest. The request object containing parameters. Returns: - any: The result of the execution. - error: An error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//   - req (*ExecutionRequest): The request object containing specific parameters.
//
// Returns:
//   - (any): The resulting any object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
func (t *CallableTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	return t.callable.Call(ctx, req)
}

// Callable - Auto-generated documentation.
//
// Summary: Callable returns the underlying Callable of the tool.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (t *CallableTool) Callable() Callable {
	return t.callable
}
