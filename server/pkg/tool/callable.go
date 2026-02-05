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
// Parameters:
//   toolDef: The definition of the tool.
//   serviceConfig: The configuration of the service the tool belongs to.
//   callable: The callable implementation for execution.
//   inputSchema: The input schema for the tool.
//   outputSchema: The output schema for the tool.
//
// Returns:
//   *CallableTool: A pointer to the created CallableTool.
//   error: An error if creation fails.
func NewCallableTool(toolDef *configv1.ToolDefinition, serviceConfig *configv1.UpstreamServiceConfig, callable Callable, inputSchema, outputSchema *structpb.Struct) (*CallableTool, error) {
	base, err := newBaseTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
	if err != nil {
		return nil, err
	}
	return &CallableTool{base}, nil
}

// Execute handles the execution of the tool.
//
// ctx is the context for the request.
// req is the request object.
//
// Returns the result.
// Returns an error if the operation fails.
func (t *CallableTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	return t.callable.Call(ctx, req)
}

// Callable returns the underlying Callable of the tool.
//
// Returns the result.
func (t *CallableTool) Callable() Callable {
	return t.callable
}
