import re

filename = 'server/pkg/upstream/grpc/grpc_test.go'

with open(filename, 'r') as f:
    content = f.read()

replacement = """
func (m *simpleMockTool) IsStreaming() bool {
	return false
}

func (m *simpleMockTool) StreamExecute(ctx context.Context, req *tool.ExecutionRequest) (<-chan any, error) {
	return nil, nil
}
"""

if "func (m *simpleMockTool) IsStreaming() bool" not in content:
    content = content.replace("type simpleMockTool struct {", replacement + "\ntype simpleMockTool struct {")

with open(filename, 'w') as f:
    f.write(content)
