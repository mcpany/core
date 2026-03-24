// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

// mockCallable is a simple mock for the Callable interface.
type mockCallable struct {
	callFunc func(ctx context.Context, req *ExecutionRequest) (any, error)
}

func (m *mockCallable) Call(ctx context.Context, req *ExecutionRequest) (any, error) {
	if m.callFunc != nil {
		return m.callFunc(ctx, req)
	}
	return nil, nil
}

// mockStreamingCallable is a mock that implements both Callable and StreamingCallable.
type mockStreamingCallable struct {
	mockCallable
	streamCallFunc func(ctx context.Context, req *ExecutionRequest) (<-chan any, error)
}

func (m *mockStreamingCallable) StreamCall(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
	if m.streamCallFunc != nil {
		return m.streamCallFunc(ctx, req)
	}
	ch := make(chan any)
	close(ch)
	return ch, nil
}

func TestNewCallableTool(t *testing.T) {
	toolDef := &configv1.ToolDefinition{
		Name: "test_tool",
	}
	serviceConfig := &configv1.UpstreamServiceConfig{}
	callable := &mockCallable{}

	t.Run("success", func(t *testing.T) {
		tool, err := NewCallableTool(toolDef, serviceConfig, callable, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, tool)
		assert.Equal(t, callable, tool.Callable())
	})

	t.Run("nil_tool_def_error", func(t *testing.T) {
		tool, err := NewCallableTool(nil, serviceConfig, callable, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, tool)
	})
}

func TestCallableTool_Execute(t *testing.T) {
	expectedResult := "test result"
	expectedError := errors.New("test error")

	toolDef := &configv1.ToolDefinition{Name: "test_tool"}
	req := &ExecutionRequest{}

	t.Run("success", func(t *testing.T) {
		callable := &mockCallable{
			callFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
				return expectedResult, nil
			},
		}
		tool, _ := NewCallableTool(toolDef, nil, callable, nil, nil)

		res, err := tool.Execute(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, res)
	})

	t.Run("error", func(t *testing.T) {
		callable := &mockCallable{
			callFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
				return nil, expectedError
			},
		}
		tool, _ := NewCallableTool(toolDef, nil, callable, nil, nil)

		res, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, res)
	})
}

func TestCallableTool_IsStreaming(t *testing.T) {
	toolDef := &configv1.ToolDefinition{Name: "test_tool"}

	t.Run("non_streaming", func(t *testing.T) {
		callable := &mockCallable{}
		tool, _ := NewCallableTool(toolDef, nil, callable, nil, nil)
		assert.False(t, tool.IsStreaming())
	})

	t.Run("streaming", func(t *testing.T) {
		callable := &mockStreamingCallable{}
		tool, _ := NewCallableTool(toolDef, nil, callable, nil, nil)
		assert.True(t, tool.IsStreaming())
	})
}

func TestCallableTool_StreamExecute(t *testing.T) {
	toolDef := &configv1.ToolDefinition{Name: "test_tool"}
	req := &ExecutionRequest{}

	t.Run("streaming_callable_success", func(t *testing.T) {
		expectedCh := make(chan any, 1)
		expectedCh <- "stream result"
		close(expectedCh)

		callable := &mockStreamingCallable{
			streamCallFunc: func(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
				return expectedCh, nil
			},
		}
		tool, _ := NewCallableTool(toolDef, nil, callable, nil, nil)

		ch, err := tool.StreamExecute(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, ch)

		res, ok := <-ch
		assert.True(t, ok)
		assert.Equal(t, "stream result", res)
	})

	t.Run("streaming_callable_error", func(t *testing.T) {
		expectedErr := errors.New("stream error")
		callable := &mockStreamingCallable{
			streamCallFunc: func(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
				return nil, expectedErr
			},
		}
		tool, _ := NewCallableTool(toolDef, nil, callable, nil, nil)

		ch, err := tool.StreamExecute(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, ch)
	})

	t.Run("fallback_non_streaming_success", func(t *testing.T) {
		expectedResult := "fallback result"
		callable := &mockCallable{
			callFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
				return expectedResult, nil
			},
		}
		tool, _ := NewCallableTool(toolDef, nil, callable, nil, nil)

		ch, err := tool.StreamExecute(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, ch)

		res, ok := <-ch
		assert.True(t, ok)
		assert.Equal(t, expectedResult, res)
	})

	t.Run("fallback_non_streaming_error", func(t *testing.T) {
		expectedErr := errors.New("fallback error")
		callable := &mockCallable{
			callFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
				return nil, expectedErr
			},
		}
		tool, _ := NewCallableTool(toolDef, nil, callable, nil, nil)

		ch, err := tool.StreamExecute(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, ch)

		res, ok := <-ch
		assert.True(t, ok)
		assert.Equal(t, expectedErr, res)
	})
}
