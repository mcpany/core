import re

filename = 'server/pkg/mcpserver/roots_tool.go'

with open(filename, 'r') as f:
    content = f.read()

replacement = """
func (t *RootsTool) IsStreaming() bool {
	return false
}

func (t *RootsTool) StreamExecute(ctx context.Context, req *tool.ExecutionRequest) (<-chan any, error) {
	return nil, nil
}
"""

if "func (t *RootsTool) IsStreaming() bool" not in content:
    content = content.replace("func (t *RootsTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {", replacement + "\nfunc (t *RootsTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {")

with open(filename, 'w') as f:
    f.write(content)

filename = 'server/pkg/upstream/openapi/openapi_test.go'

with open(filename, 'r') as f:
    content = f.read()

replacement = """
func (m *MockTool) IsStreaming() bool {
	return false
}

func (m *MockTool) StreamExecute(ctx context.Context, req *tool.ExecutionRequest) (<-chan any, error) {
	return nil, nil
}
"""

if "func (m *MockTool) IsStreaming() bool" not in content:
    content = content.replace("func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {", replacement + "\nfunc (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {")

with open(filename, 'w') as f:
    f.write(content)
