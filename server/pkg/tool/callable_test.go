package tool
import (
	"context"
	"testing"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)
type mockCallable struct {
	callFunc func(ctx context.Context, req *ExecutionRequest) (any, error)
}
func (m *mockCallable) Call(ctx context.Context, req *ExecutionRequest) (any, error) {
	if m.callFunc != nil { return m.callFunc(ctx, req) }
	return nil, nil
}
func (m *mockCallable) IsStreaming() bool { return false }
func (m *mockCallable) StreamCall(ctx context.Context, req *ExecutionRequest) (<-chan any, error) { return nil, nil }
func TestNewCallableTool(t *testing.T) {
	toolDef := configv1.ToolDefinition_builder{Name: proto.String("test_tool")}.Build()
	serviceConfig := &configv1.UpstreamServiceConfig{}
	callable := &mockCallable{}
	t.Run("success", func(t *testing.T) {
		tool, err := NewCallableTool(toolDef, serviceConfig, callable, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, tool)
		assert.Equal(t, callable, tool.Callable())
	})
}
