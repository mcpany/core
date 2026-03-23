import re

# roots_tool.go uses `_ *tool.ExecutionRequest`
filepath = 'server/pkg/mcpserver/roots_tool.go'
with open(filepath, 'r') as f:
    content = f.read()

content = content.replace('func (t *RootsTool) Execute(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {', '''// Execute executes the "mcp:list_roots" tool.
//
// Summary: Executes the roots tool.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - _ (*tool.ExecutionRequest): The execution request (unused).
//
// Returns:
//   - any: The roots listing.
//   - error: An error if roots are unavailable.
//
// Errors:
//   - Returns error on retrieval failure.
//
// Side Effects:
//   - Queries the connected client for its roots.
func (t *RootsTool) Execute(ctx context.Context, _ *tool.ExecutionRequest) (any, error) {''')

with open(filepath, 'w') as f:
    f.write(content)


# mock_tool.go uses `StreamExecute(ctx context.Context, req *ExecutionRequest)` for MockTool
# and `*tool.ExecutionRequest` depending on the file. Let's patch specifically for mock_tool.go

filepath = 'server/pkg/tool/mock_tool.go'
with open(filepath, 'r') as f:
    content = f.read()

content = content.replace('func (m *MockTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {', '''// StreamExecute simulates streaming execution for testing.
//
// Summary: Simulates streaming execution.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*ExecutionRequest): The execution request.
//
// Returns:
//   - <-chan any: A mock streaming channel.
//   - error: An error if configured to fail.
//
// Errors:
//   - Returns an error if mock fails.
//
// Side Effects:
//   - None.
func (m *MockTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {''')

with open(filepath, 'w') as f:
    f.write(content)
