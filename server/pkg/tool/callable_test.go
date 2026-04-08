// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStreamingCallable struct{}

func (m *mockStreamingCallable) Call(ctx context.Context, req *ExecutionRequest) (any, error) {
	return "called", nil
}

func (m *mockStreamingCallable) StreamCall(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
	ch := make(chan any, 1)
	ch <- "stream-called"
	close(ch)
	return ch, nil
}

func TestCallableTool_New(t *testing.T) {
	t.Parallel()

	toolDef := &configv1.ToolDefinition{
		Name: "test-callable",
	}

	callable := &mockCallable{} // from base_test.go
	ct, err := NewCallableTool(toolDef, nil, callable, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, ct)

	assert.Same(t, callable, ct.Callable())

	// Test IsStreaming for non-streaming
	assert.False(t, ct.IsStreaming())

	// Test Execute
	res, err := ct.Execute(context.Background(), &ExecutionRequest{})
	assert.NoError(t, err)
	assert.Nil(t, res) // mockCallable returns nil, nil
}

func TestCallableTool_Streaming(t *testing.T) {
	t.Parallel()

	toolDef := &configv1.ToolDefinition{
		Name: "test-streaming-callable",
	}

	streamingCallable := &mockStreamingCallable{}
	ct, err := NewCallableTool(toolDef, nil, streamingCallable, nil, nil)
	require.NoError(t, err)

	assert.True(t, ct.IsStreaming())

	ch, err := ct.StreamExecute(context.Background(), &ExecutionRequest{})
	require.NoError(t, err)

	val := <-ch
	assert.Equal(t, "stream-called", val)
}

func TestCallableTool_NonStreamingFallback(t *testing.T) {
	t.Parallel()

	toolDef := &configv1.ToolDefinition{
		Name: "test-callable",
	}

	// mockCallable is non-streaming
	callable := &mockCallable{}
	ct, err := NewCallableTool(toolDef, nil, callable, nil, nil)
	require.NoError(t, err)

	assert.False(t, ct.IsStreaming())

	ch, err := ct.StreamExecute(context.Background(), &ExecutionRequest{})
	require.NoError(t, err)

	val := <-ch
	assert.Nil(t, val) // mockCallable returns nil
}
