package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCallableTool(t *testing.T) {
	toolDef := &configv1.ToolDefinition{Name: "test"}
	callable := &dummyCallable{}
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{})
	outputSchema, _ := structpb.NewStruct(map[string]interface{}{})

	ct, err := NewCallableTool(toolDef, nil, callable, inputSchema, outputSchema)
	require.NoError(t, err)

	assert.Equal(t, callable, ct.Callable())

	res, err := ct.Execute(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "dummy", res)

	assert.False(t, ct.IsStreaming())

	ch, err := ct.StreamExecute(context.Background(), nil)
	require.NoError(t, err)
	assert.NotNil(t, ch)
	resStream := <-ch
	assert.Equal(t, "dummy", resStream)
}

type streamingCallable struct {
	dummyCallable
}

func (s *streamingCallable) StreamCall(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
	ch := make(chan any, 1)
	ch <- "stream_dummy"
	close(ch)
	return ch, nil
}

func TestCallableTool_Streaming(t *testing.T) {
	toolDef := &configv1.ToolDefinition{Name: "stream_test"}
	callable := &streamingCallable{}
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{})
	outputSchema, _ := structpb.NewStruct(map[string]interface{}{})

	ct, err := NewCallableTool(toolDef, nil, callable, inputSchema, outputSchema)
	require.NoError(t, err)

	assert.True(t, ct.IsStreaming())

	ch, err := ct.StreamExecute(context.Background(), nil)
	require.NoError(t, err)
	assert.NotNil(t, ch)
	resStream := <-ch
	assert.Equal(t, "stream_dummy", resStream)
}
