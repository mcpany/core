import re
import os

files_to_patch = [
    'server/pkg/app/api_alignment.go',
    'server/pkg/middleware/aia_broker.go',
    'server/pkg/mcpserver/roots_tool.go',
    'server/pkg/upstream/sql/tool.go',
    'server/pkg/tool/base.go',
    'server/pkg/tool/webrtc.go',
    'server/pkg/tool/websocket.go',
    'server/pkg/tool/types.go',
    'server/pkg/tool/mock_tool.go'
]

replacements = {
    # pkg/app/api_alignment.go
    '// SubagentStatus defines the structure for AIA heartbeat responses.': '''// SubagentStatus defines the structure for AIA heartbeat responses.
//
// Summary: Defines the structure for AIA heartbeat responses.''',

    # pkg/middleware/aia_broker.go
    '''// NewAIABroker creates a new AIABroker middleware instance.
//
// Parameters:''': '''// NewAIABroker creates a new AIABroker middleware instance.
//
// Summary: Creates a new AIA Broker middleware instance.
//
// Parameters:''',

    '''// Execute enforces intent alignment before proceeding to the next handler.
//
// Parameters:''': '''// Execute enforces intent alignment before proceeding to the next handler.
//
// Summary: Enforces intent alignment before proceeding to the next handler.
//
// Parameters:'''
}

for filepath in files_to_patch:
    if not os.path.exists(filepath):
        continue
    with open(filepath, 'r') as f:
        content = f.read()

    for old, new in replacements.items():
        content = content.replace(old, new)

    # For RootsTool, SQLTool, WebsocketTool, WebrtcTool, MockTool, BaseTool, GRPCTool, HTTPTool, MCPTool, OpenAPITool, CommandTool, LocalCommandTool
    # Replace IsStreaming, StreamExecute, Execute

    content = re.sub(r'func \(\w+ \*\w+\) IsStreaming\(\) bool \{', r'''// IsStreaming returns whether the tool supports streaming execution.
//
// Summary: Returns whether streaming is supported.
//
// Parameters:
//   - None.
//
// Returns:
//   - bool: True if streaming is supported, otherwise false.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
\g<0>''', content)

    content = re.sub(r'func \(\w+ \*\w+\) StreamExecute\(ctx context\.Context, req \*tool\.ExecutionRequest\) \(<-chan any, error\) \{', r'''// StreamExecute executes the tool in streaming mode.
//
// Summary: Executes the tool in streaming mode.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*tool.ExecutionRequest): The execution request.
//
// Returns:
//   - <-chan any: The channel of stream results.
//   - error: An error if streaming initialization fails.
//
// Errors:
//   - Returns error if the underlying stream cannot be established.
//
// Side Effects:
//   - Begins asynchronous processing and emitting to the returned channel.
\g<0>''', content)

    content = re.sub(r'func \(\w+ \*\w+\) Execute\(ctx context\.Context, req \*tool\.ExecutionRequest\) \(any, error\) \{', r'''// Execute runs the tool with the given request.
//
// Summary: Executes the tool.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*tool.ExecutionRequest): The execution request parameters.
//
// Returns:
//   - any: The execution result.
//   - error: An error if execution fails.
//
// Errors:
//   - Returns an error on operational failures.
//
// Side Effects:
//   - May invoke upstream services or mutate state depending on the tool logic.
\g<0>''', content)

    content = re.sub(r'func \(\w+ \*\w+\) Execute\(ctx context\.Context, req \*ExecutionRequest\) \(any, error\) \{', r'''// Execute runs the tool with the given request.
//
// Summary: Executes the tool.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*ExecutionRequest): The execution request parameters.
//
// Returns:
//   - any: The execution result.
//   - error: An error if execution fails.
//
// Errors:
//   - Returns an error on operational failures.
//
// Side Effects:
//   - May invoke upstream services or mutate state depending on the tool logic.
\g<0>''', content)

    content = re.sub(r'func \(\w+ \*\w+\) GetCacheConfig\(\) \*configv1\.CacheConfig \{', r'''// GetCacheConfig retrieves the cache configuration for the tool.
//
// Summary: Retrieves the cache configuration.
//
// Parameters:
//   - None.
//
// Returns:
//   - *configv1.CacheConfig: The configuration or nil.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
\g<0>''', content)

    with open(filepath, 'w') as f:
        f.write(content)
