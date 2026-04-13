// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// mockCallable implements Callable for testing.
type mockCallable struct {
	called bool
	req    *ExecutionRequest
	result any
	err    error
}

func (m *mockCallable) Call(ctx context.Context, req *ExecutionRequest) (any, error) {
	m.called = true
	m.req = req
	return m.result, m.err
}

// mockStreamingCallable implements StreamingCallable for testing.
type mockStreamingCallable struct {
	mockCallable
	streamResult <-chan any
	streamErr    error
}

func (m *mockStreamingCallable) StreamCall(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
	m.called = true
	m.req = req
	return m.streamResult, m.streamErr
}

func ptrStr(s string) *string {
	return &s
}

func TestBaseToolExecution(t *testing.T) {
	t.Parallel()

	inputSchema, _ := structpb.NewStruct(map[string]any{"type": "object"})
	outputSchema, _ := structpb.NewStruct(map[string]any{"type": "object"})

	toolDef := configv1.ToolDefinition_builder{
		Name: ptrStr("test_base_tool"),
		Description: ptrStr("A base tool for testing"),
	}.Build()

	callable := &mockCallable{}
	base, err := newBaseTool(toolDef, nil, callable, inputSchema, outputSchema)
	require.NoError(t, err)
	require.NotNil(t, base)

	// Test Tool()
	pbTool := base.Tool()
	require.NotNil(t, pbTool)

	// Test MCPTool() - lazy initialization
	mcpTool := base.MCPTool()
	require.NotNil(t, mcpTool)
	assert.Equal(t, ".test_base_tool", mcpTool.Name) // Appears to prefix with .

	// Ensure calling MCPTool again returns the same instance (sync.Once works)
	mcpTool2 := base.MCPTool()
	assert.Same(t, mcpTool, mcpTool2)

	// Test GetCacheConfig()
	assert.Nil(t, base.GetCacheConfig())

	// Test IsStreaming()
	assert.False(t, base.IsStreaming())

	// Test StreamExecute()
	ch, streamErr := base.StreamExecute(context.Background(), nil)
	assert.Nil(t, ch)
	assert.NoError(t, streamErr)
}

func TestCallableToolExecution(t *testing.T) {
	t.Parallel()

	toolDef := configv1.ToolDefinition_builder{
		Name: ptrStr("test_callable_tool"),
		Description: ptrStr("A callable tool for testing"),
	}.Build()

	inputSchema, _ := structpb.NewStruct(map[string]any{"type": "object"})
	outputSchema, _ := structpb.NewStruct(map[string]any{"type": "object"})

	t.Run("Non-streaming Callable", func(t *testing.T) {
		callable := &mockCallable{result: "success"}
		ct, err := NewCallableTool(toolDef, nil, callable, inputSchema, outputSchema)
		require.NoError(t, err)
		require.NotNil(t, ct)

		// Test Callable()
		assert.Same(t, callable, ct.Callable())

		// Test IsStreaming()
		assert.False(t, ct.IsStreaming())

		// Test Execute()
		req := &ExecutionRequest{}
		res, err := ct.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "success", res)
		assert.True(t, callable.called)
		assert.Same(t, req, callable.req)

		// Test StreamExecute() fallback
		ch, err := ct.StreamExecute(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, ch)

		// Read from the channel to ensure the fallback works
		streamRes, ok := <-ch
		require.True(t, ok)
		assert.Equal(t, "success", streamRes)

		// Channel should be closed
		_, ok = <-ch
		assert.False(t, ok)
	})

	t.Run("Streaming Callable", func(t *testing.T) {
		streamCh := make(chan any, 1)
		streamCh <- "stream_success"
		close(streamCh)

		callable := &mockStreamingCallable{
			mockCallable: mockCallable{result: "sync_success"},
			streamResult: streamCh,
			streamErr:    nil,
		}

		ct, err := NewCallableTool(toolDef, nil, callable, inputSchema, outputSchema)
		require.NoError(t, err)
		require.NotNil(t, ct)

		// Test IsStreaming()
		assert.True(t, ct.IsStreaming())

		// Test StreamExecute() uses StreamCall
		req := &ExecutionRequest{}
		ch, err := ct.StreamExecute(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, ch)
		assert.Equal(t, (<-chan any)(streamCh), ch)
		assert.True(t, callable.called)
	})
}
