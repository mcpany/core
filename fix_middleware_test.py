import re

filename = 'server/pkg/middleware/call_policy_test.go'

with open(filename, 'r') as f:
    content = f.read()

replacement = """
func (m *callPolicyMockTool) IsStreaming() bool {
	return false
}

func (m *callPolicyMockTool) StreamExecute(ctx context.Context, req *tool.ExecutionRequest) (<-chan any, error) {
	return nil, nil
}
"""

if "func (m *callPolicyMockTool) IsStreaming() bool" not in content:
    content = content.replace("type callPolicyMockTool struct {", replacement + "\ntype callPolicyMockTool struct {")

with open(filename, 'w') as f:
    f.write(content)

filename = 'server/pkg/middleware/ratelimit_test.go'

with open(filename, 'r') as f:
    content = f.read()

replacement = """
func (m *rateLimitMockTool) IsStreaming() bool {
	return false
}

func (m *rateLimitMockTool) StreamExecute(ctx context.Context, req *tool.ExecutionRequest) (<-chan any, error) {
	return nil, nil
}
"""

if "func (m *rateLimitMockTool) IsStreaming() bool" not in content:
    content = content.replace("type rateLimitMockTool struct {", replacement + "\ntype rateLimitMockTool struct {")

with open(filename, 'w') as f:
    f.write(content)
