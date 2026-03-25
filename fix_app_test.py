import re

filename = 'server/pkg/app/api_handlers_extra_test.go'

with open(filename, 'r') as f:
    content = f.read()

replacement = """
func (t *TestMockTool) IsStreaming() bool {
	return false
}

func (t *TestMockTool) StreamExecute(ctx context.Context, req *tool.ExecutionRequest) (<-chan any, error) {
	return nil, nil
}
"""

if "func (t *TestMockTool) IsStreaming() bool" not in content:
    content = content.replace("type TestMockTool struct {", replacement + "\ntype TestMockTool struct {")

with open(filename, 'w') as f:
    f.write(content)
