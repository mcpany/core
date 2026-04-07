// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// Summary: CallableTool represents a data structure.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
type CallableTool struct {
	*baseTool
}

// NewCallableTool creates a new CallableTool.
//
// Summary: Creates a new tool that wraps a Callable interface.
// Summary: NewCallableTool executes the operation.
//
// Parameters:
//   - toolDef *configv1.ToolDefinition: Input parameter.
//   - serviceConfig *configv1.UpstreamServiceConfig: Input parameter.
//   - callable Callable: Input parameter.
//   - inputSchema: Input parameter.
//   - outputSchema *structpb.Struct: Input parameter.
//
// Returns:
//   - (*CallableTool, error): Result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None
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
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request object containing parameters.
//
// Returns:
//   - any: The result of the execution.
//   - error: An error if the operation fails.
// Summary: Execute executes the operation.
//
// Parameters:
//   - ctx context.Context: Input parameter.
//   - req *ExecutionRequest: Input parameter.
//
// Returns:
//   - (any, error): Result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None
func (t *CallableTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	return t.callable.Call(ctx, req)
}

// Callable returns the underlying Callable of the tool.
//
// Summary: Retrieves the underlying Callable interface.
//
// Returns:
//   - Callable: The underlying callable.
// Summary: Callable executes the operation.
//
// Parameters:
//   - None
//
// Returns:
//   - Callable {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (t *CallableTool) Callable() Callable {
	return t.callable
}

// IsStreaming returns true if the underlying callable supports streaming.
//
// Summary: Checks if the tool supports streaming execution.
//
// Returns:
//   - bool: True if streaming is supported.
// Summary: IsStreaming executes the operation.
//
// Parameters:
//   - None
//
// Returns:
//   - bool {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (t *CallableTool) IsStreaming() bool {
	_, ok := t.callable.(StreamingCallable)
	return ok
}

// StreamExecute handles the streaming execution of the tool.
//
// Summary: Executes the underlying callable in streaming mode.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request object containing parameters.
//
// Returns:
//   - <-chan any: A channel that emits streaming results.
// Summary: StreamExecute executes the operation.
//
// Parameters:
//   - ctx context.Context: Input parameter.
//   - req *ExecutionRequest: Input parameter.
//
// Returns:
//   - (<-chan any, error): Result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None
func (t *CallableTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
	if sc, ok := t.callable.(StreamingCallable); ok {
		return sc.StreamCall(ctx, req)
	}
	// Fallback to non-streaming execution and push to a single-item channel
	ch := make(chan any, 1)
	go func() {
		defer close(ch)
		res, err := t.Execute(ctx, req)
		if err != nil {
			ch <- err
		} else {
			ch <- res
		}
	}()
	return ch, nil
}
