import re

filename = 'server/pkg/serviceregistry/registry_test.go'

with open(filename, 'r') as f:
    content = f.read()

replacement = """
func (m *mockTool) IsStreaming() bool {
	return false
}

func (m *mockTool) StreamExecute(ctx context.Context, req *tool.ExecutionRequest) (<-chan any, error) {
	return nil, nil
}
"""

if "func (m *mockTool) IsStreaming() bool" not in content:
    content = content.replace("type mockTool struct {", replacement + "\ntype mockTool struct {")

with open(filename, 'w') as f:
    f.write(content)
