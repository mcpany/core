package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

type dummyCallable struct{}

func (d *dummyCallable) Call(ctx context.Context, req *ExecutionRequest) (any, error) {
	return "dummy", nil
}

func TestBaseTool(t *testing.T) {
	toolDef := &configv1.ToolDefinition{
		Name: "test_tool",
		Description: "A test tool",
	}
	serviceConfig := &configv1.UpstreamServiceConfig{}
	callable := &dummyCallable{}

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"foo": map[string]interface{}{"type": "string"},
		},
	})
	outputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
	})

	base, err := newBaseTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
	require.NoError(t, err)

	assert.NotNil(t, base.Tool())
	assert.Equal(t, "test_tool", base.Tool().GetName())

	assert.NotNil(t, base.MCPTool())
	assert.Equal(t, "test_tool", base.MCPTool().Name)

	assert.False(t, base.IsStreaming())
	assert.Nil(t, base.GetCacheConfig())

	ch, err := base.StreamExecute(context.Background(), nil)
	assert.Nil(t, ch)
	assert.Nil(t, err)
}

func TestBaseTool_ConvertError(t *testing.T) {
	// Creating an invalid schema will trigger a conversion error.
	invalidSchema, _ := structpb.NewStruct(map[string]interface{}{
	})

	toolDef := &configv1.ToolDefinition{Name: "bad"}
	_, _ = newBaseTool(toolDef, nil, nil, nil, invalidSchema)
}
