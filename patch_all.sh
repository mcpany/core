#!/bin/bash
set -e

# Create docs
mkdir -p docs/research
cat << 'MD_EOF' > docs/research/2026-06-18_streaming_support.md
# Strategic Evolution: 2026-06-18

### Focus: Real-Time Telemetry and Streaming Execution Results
**Context**: As autonomous swarms perform deeper and more long-running tasks, waiting for an entire tool execution to complete before receiving output causes "Cognitive Stall" for the user and parent agents. Long-running shell commands, streaming LLM outputs, or continuous metric gathering require a mechanism to stream results incrementally rather than returning a monolithic block at the end. Current architecture buffers output, leading to delayed feedback loops.
**Strategic Pivot**:
- **Streaming Tool Execution**: MCP Any will evolve to support real-time streaming of tool outputs. We will introduce `StreamExecutionRequest` and `StreamingCallable` interfaces that allow tools to emit progress chunks, enabling faster feedback for interactive UI and downstream agents.
- **Incremental Context Updates**: By streaming tool execution, parent agents can dynamically adjust or abort long-running tasks if early streaming results deviate from the mission root, mitigating context pollution early.

## Core Logic

The core logic introduces an `EventEmitter` interface within `tool.ExecutionRequest` or introduces a new `StreamingCallable` that accepts a channel or an emitter function.

\`\`\`mermaid
graph TD
    User[User / AI Agent] -->|Tool Execution| Gateway[MCP Any Gateway]
    Gateway -->|Execute| Adapter[Upstream Adapter]
    Adapter -->|Initial Setup| Subprocess[Long Running Process]
    Subprocess --o|Stream Chunk 1| Adapter
    Adapter --o|Stream Chunk 1| Gateway
    Gateway --o|Update| User
    Subprocess --o|Stream Chunk 2| Adapter
    Adapter --o|Stream Chunk 2| Gateway
    Gateway --o|Update| User
    Subprocess -->|Exit Status| Adapter
    Adapter -->|Final Result| Gateway
    Gateway -->|Final Result| User
\`\`\`

This ensures that partial results can be streamed.
MD_EOF

# Patch types.go
cat << 'PY_EOF' > patch_types.py
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

with open(filename, 'w') as f:
    f.write(content)

PY_EOF
python3 patch_types.py


cat << 'PY_EOF' > patch_base.py
import re

filename = 'server/pkg/tool/base.go'
with open(filename, 'r') as f:
    content = f.read()

if "context" not in content:
    content = content.replace('import (\n\t"sync"', 'import (\n\t"context"\n\t"sync"')

if "IsStreaming() bool" not in content:
    content = content.replace("func (t *baseTool) GetCacheConfig() *configv1.CacheConfig {", """// IsStreaming returns true if the tool supports streaming execution.
//
// Summary: Checks if the tool supports streaming execution.
//
// Returns:
//   - bool: True if streaming is supported.
func (t *baseTool) IsStreaming() bool {
	return false
}

// StreamExecute handles the streaming execution of the tool.
//
// Summary: Executes the tool in streaming mode.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request payload.
//
// Returns:
//   - <-chan any: A channel that emits streaming results.
//   - error: An error if the operation fails or streaming is not supported.
func (t *baseTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
	return nil, nil // Should be implemented by embedding struct if supported
}

func (t *baseTool) GetCacheConfig() *configv1.CacheConfig {""")

with open(filename, 'w') as f:
    f.write(content)
PY_EOF
python3 patch_base.py

cat << 'PY_EOF' > patch_callable.py
import re

filename = 'server/pkg/tool/callable.go'
with open(filename, 'r') as f:
    content = f.read()

if "IsStreaming() bool" not in content:
    content = content.replace("func (t *CallableTool) Callable() Callable {\n\treturn t.callable\n}", """func (t *CallableTool) Callable() Callable {
	return t.callable
}

// IsStreaming returns true if the underlying callable supports streaming.
//
// Summary: Checks if the tool supports streaming execution.
//
// Returns:
//   - bool: True if streaming is supported.
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
//   - error: An error if the operation fails or streaming is not supported.
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
}""")

with open(filename, 'w') as f:
    f.write(content)
PY_EOF
python3 patch_callable.py

cat << 'PY_EOF' > patch_tool_types.py
import re

files_to_patch = [
    'server/pkg/tool/types.go',
    'server/pkg/tool/websocket.go',
    'server/pkg/tool/webrtc.go',
    'server/pkg/tool/extra_management_test.go',
    'server/pkg/tool/fuzzy_test.go',
    'server/pkg/tool/mock_tool.go',
    'server/pkg/tool/management_coverage_test.go',
    'server/pkg/serviceregistry/registry_test.go',
    'server/pkg/upstream/sql/tool.go',
    'server/pkg/resource/dynamic_resource_test.go',
    'server/pkg/upstream/grpc/grpc_test.go'
]

types = ["GRPCTool", "HTTPTool", "MCPTool", "OpenAPITool", "CommandTool", "LocalCommandTool", "WebsocketTool", "WebrtcTool", "mockToolSimple", "MockTool", "mockTool", "Tool", "simpleMockTool"]

for filename in files_to_patch:
    try:
        with open(filename, 'r') as f:
            content = f.read()

        for t in types:
            if t in content and "func (t *"+t+") IsStreaming() bool" not in content and "func (m *"+t+") IsStreaming() bool" not in content:
                # We can replace Execute
                if f"func (t *{t}) Execute" in content:
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

func (t *{t}) Execute"""
                    content = content.replace(f"func (t *{t}) Execute", replacement)
                elif f"func (m *{t}) Execute" in content:
                    prefix = "tool." if ("ExecutionRequest" not in content and "tool.ExecutionRequest" in content) else ""
                    # Check what the actual signature uses
                    exec_sig = ""
                    if f"func (m *{t}) Execute(ctx context.Context, req *ExecutionRequest)" in content:
                        exec_sig = "*ExecutionRequest"
                    elif f"func (m *{t}) Execute(ctx context.Context, req *tool.ExecutionRequest)" in content:
                        exec_sig = "*tool.ExecutionRequest"
                    else:
                        continue

                    replacement = f"""func (m *{t}) IsStreaming() bool {{
	return false
}}

func (m *{t}) StreamExecute(ctx context.Context, req {exec_sig}) (<-chan any, error) {{
	ch := make(chan any, 1)
	go func() {{
		defer close(ch)
		res, err := m.Execute(ctx, req)
		if err != nil {{
			ch <- err
		}} else {{
			ch <- res
		}}
	}}()
	return ch, nil
}}

func (m *{t}) Execute"""
                    content = content.replace(f"func (m *{t}) Execute", replacement)

        with open(filename, 'w') as f:
            f.write(content)
    except Exception as e:
        pass
PY_EOF
python3 patch_tool_types.py
