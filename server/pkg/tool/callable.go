// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// CallableTool implements the Tool interface for a tool that is executed by a
//
// Summary: Implements the Tool interface for a tool that is executed by a
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

// NewCallableTool creates a new CallableTool.
//
// Summary: Creates a new CallableTool.
//
// Parameters:
//   - toolDef (*configv1.ToolDefinition): Parameter.
//   - serviceConfig (*configv1.UpstreamServiceConfig): Parameter.
//   - callable (Callable): Parameter.
//   - inputSchema (*structpb.Struct): Parameter.
//   - outputSchema (*structpb.Struct): Parameter.
//
// Returns:
//   - *CallableTool: Return value.
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
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

// Execute handles the execution of the tool.
//
// Summary: Handles the execution of the tool.
//
// Parameters:
//   - ctx (context.Context): Parameter.
//   - req (*ExecutionRequest): Parameter.
//
// Returns:
//   - any: Return value.
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

func (t *CallableTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	return t.callable.Call(ctx, req)
}

// Callable returns the underlying of the tool.
//
// Summary: Returns the underlying of the tool.
//
// Parameters:
//   - None.
//
// Returns:
//   - Callable: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (t *CallableTool) Callable() Callable {
	return t.callable
}

// IsStreaming returns true if the underlying callable supports streaming.
//
// Summary: Returns true if the underlying callable supports streaming.
//
// Parameters:
//   - None.
//
// Returns:
//   - bool: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func (t *CallableTool) IsStreaming() bool {
	_, ok := t.callable.(StreamingCallable)
	return ok
}

// StreamExecute handles the streaming execution of the tool.
//
// Summary: Handles the streaming execution of the tool.
//
// Parameters:
//   - ctx (context.Context): Parameter.
//   - req (*ExecutionRequest): Parameter.
//
// Returns:
//   - <-chan any: Return value.
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

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
