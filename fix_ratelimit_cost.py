import re

filename = 'server/pkg/middleware/ratelimit_cost_test.go'

with open(filename, 'r') as f:
    content = f.read()

replacement = """
func (m *MockToolForCost) IsStreaming() bool {
	return false
}

func (m *MockToolForCost) StreamExecute(ctx context.Context, req *tool.ExecutionRequest) (<-chan any, error) {
	return nil, nil
}
"""

if "func (m *MockToolForCost) IsStreaming() bool" not in content:
    content = content.replace("type MockToolForCost struct {", replacement + "\ntype MockToolForCost struct {")

with open(filename, 'w') as f:
    f.write(content)
