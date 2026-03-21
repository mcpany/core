import re

filename = 'server/pkg/topology/manager_test.go'

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
    content = content.replace("type MockTool struct {", replacement + "\ntype MockTool struct {")

with open(filename, 'w') as f:
    f.write(content)
