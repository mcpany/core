// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

// NewCallableTool creates a new CallableTool.
//
// Summary: Creates a new tool that wraps a Callable interface.
//
// Parameters:
//   - toolDef: *configv1.ToolDefinition. The definition of the tool.
//   - serviceConfig: *configv1.UpstreamServiceConfig. The configuration of the service the tool belongs to.
//   - callable: Callable. The callable implementation for execution.
//   - inputSchema: *structpb.Struct. The input schema for the tool.
//   - outputSchema: *structpb.Struct. The output schema for the tool.
// Execute handles the execution of the tool.
//
// Summary: Executes the underlying callable.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request object containing parameters.
//
// Callable returns the underlying Callable of the tool.
//
// StreamExecute handles the streaming execution of the tool.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Summary: Executes the underlying callable in streaming mode.
//
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request object containing parameters.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - <-chan any: A channel that emits streaming results.
//   - error: An error if the operation fails or streaming is not supported.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
//   - None.
// Side Effects:
//   - None.
// Errors:
//   - None.
// Returns:
// Execute handles the execution of the tool.
//   - outputSchema: *structpb.Struct. The output schema for the tool.
//   - inputSchema: *structpb.Struct. The input schema for the tool.
//   - callable: Callable. The callable implementation for execution.
//   - serviceConfig: *configv1.UpstreamServiceConfig. The configuration of the service the tool belongs to.
//   - toolDef: *configv1.ToolDefinition. The definition of the tool.
// Parameters:
//
// Summary: Creates a new tool that wraps a Callable interface.
//
// NewCallableTool creates a new CallableTool.
		} else {
			ch <- res
		}
	}()
	return ch, nil
}
