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
