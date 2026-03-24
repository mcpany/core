import re

filename = 'server/pkg/tool/base.go'
with open(filename, 'r') as f:
    content = f.read()

if "context" not in content:
    content = content.replace('import (\n\t"sync"', 'import (\n\t"context"\n\t"sync"')

if "IsStreaming() bool" not in content:
    content = content.replace("func (t *baseTool) GetCacheConfig() *configv1.CacheConfig {", """// IsStreaming returns true if the tool supports streaming execution.
//
// Summary: Checks if the tool supports streaming execution.
//
// Returns:
//   - bool: True if streaming is supported.
func (t *baseTool) IsStreaming() bool {
	return false
}

// StreamExecute handles the streaming execution of the tool.
//
// Summary: Executes the tool in streaming mode.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *ExecutionRequest. The request payload.
//
// Returns:
//   - <-chan any: A channel that emits streaming results.
//   - error: An error if the operation fails or streaming is not supported.
func (t *baseTool) StreamExecute(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
	return nil, nil // Should be implemented by embedding struct if supported
}

func (t *baseTool) GetCacheConfig() *configv1.CacheConfig {""")

with open(filename, 'w') as f:
    f.write(content)
