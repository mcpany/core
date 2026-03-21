import re

filename = 'server/pkg/tool/types.go'
with open(filename, 'r') as f:
    content = f.read()

# Add IsStreaming and StreamExecute to Tool interface
if "IsStreaming() bool" not in content:
    content = content.replace("Execute(ctx context.Context, req *ExecutionRequest) (any, error)", """Execute(ctx context.Context, req *ExecutionRequest) (any, error)

	// IsStreaming returns true if the tool supports streaming execution.
	//
	// Summary: Checks if the tool supports streaming execution.
	//
	// Returns:
	//   - bool: True if streaming is supported.
	IsStreaming() bool

	// StreamExecute runs the tool in streaming mode, returning a channel of results.
	//
	// Summary: Executes the tool in streaming mode.
	//
	// Parameters:
	//   - ctx: context.Context. The execution context.
	//   - req: *ExecutionRequest. The request payload.
	//
	// Returns:
	//   - <-chan any: A channel that emits streaming results.
	//   - error: An error if the operation fails or streaming is not supported.
	//
	// Side Effects:
	//   - Executes the underlying service logic in a streaming manner.
	StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error)""")

# Add StreamingCallable interface
if "StreamingCallable" not in content:
    content = content.replace("Call(ctx context.Context, req *ExecutionRequest) (any, error)\n}", """Call(ctx context.Context, req *ExecutionRequest) (any, error)
}

// StreamingCallable is an interface that represents a callable tool that can stream output.
//
// Summary: Interface for executing a tool with streaming output.
type StreamingCallable interface {
	Callable

	// StreamCall executes the callable with the given request, emitting updates to the channel.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - req: *ExecutionRequest. The execution request details.
	//
	// Returns:
	//   - <-chan any: A channel that emits streaming results.
	//   - error: An error if the initial operation fails.
	StreamCall(ctx context.Context, req *ExecutionRequest) (<-chan any, error)
}""")

types = ["GRPCTool", "HTTPTool", "MCPTool", "OpenAPITool", "CommandTool", "LocalCommandTool"]

for t in types:
    if f"func (t *{t}) IsStreaming() bool" not in content:
        pattern = r"(func \(t \*" + t + r"\) Execute\(ctx context\.Context, req \*ExecutionRequest\) \(any, error\) {)"
        replacement = f"""// IsStreaming returns true if the tool supports streaming.
//
// Summary: Checks if the tool supports streaming execution.
//
// Returns:
//   - bool: True if streaming is supported.
func (t *{t}) IsStreaming() bool {{
	return false
}}

// StreamExecute executes the tool in streaming mode.
//
// Summary: Executes the tool in streaming mode.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request object containing parameters.
//
// Returns:
//   - <-chan any: A channel that emits streaming results.
//   - error: An error if the operation fails or streaming is not supported.
func (t *{t}) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {{
	ch := make(chan any, 1)
	go func() {{
		defer close(ch)
		res, err := t.Execute(ctx, req)
		if err != nil {{
			ch <- err
		}} else {{
			ch <- res
		}}
	}}()
	return ch, nil
}}

\\1"""

        # In types.go, the issue was that the regex replaced the comment containing "func (t *GRPCTool) Execute..." before the function definition.
        # Wait, no. The issue was that replacing `Execute(ctx context.Context, req *ExecutionRequest) (any, error)`
        # replaced EVERY occurrence of `Execute` instead of just the one inside `type Tool interface`.
        pass


with open(filename, 'w') as f:
    f.write(content)
