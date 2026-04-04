// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type mockStreamingCallable struct {
	mockCallable
}

func (m *mockStreamingCallable) StreamCall(ctx context.Context, req *ExecutionRequest) (<-chan any, error) {
	ch := make(chan any, 1)
	ch <- "streaming-mock"
	close(ch)
	return ch, nil
}

func TestCallableTool_New(t *testing.T) {
	toolDef := &configv1.ToolDefinition{
		Name:        proto.String("test_tool"),
		Description: proto.String("A test tool"),
		ServiceId:   proto.String("service-id"),
	}

	serviceConfig := &configv1.UpstreamServiceConfig{
		Id: proto.String("service-id"),
	}

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
	})
	outputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
	})

	t.Run("NewCallableTool", func(t *testing.T) {
		callable := &mockCallable{}
		tool, err := NewCallableTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		require.NoError(t, err)
		assert.NotNil(t, tool)
		assert.Equal(t, callable, tool.Callable())
	})

	t.Run("Execute", func(t *testing.T) {
		callable := &mockCallable{}
		tool, err := NewCallableTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		require.NoError(t, err)

		res, err := tool.Execute(context.Background(), &ExecutionRequest{})
		require.NoError(t, err)
		assert.Equal(t, "mock", res)
	})

	t.Run("IsStreaming_False", func(t *testing.T) {
		callable := &mockCallable{}
		tool, err := NewCallableTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		require.NoError(t, err)

		assert.False(t, tool.IsStreaming())
	})

	t.Run("IsStreaming_True", func(t *testing.T) {
		callable := &mockStreamingCallable{}
		tool, err := NewCallableTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		require.NoError(t, err)

		assert.True(t, tool.IsStreaming())
	})

	t.Run("StreamExecute_NonStreaming", func(t *testing.T) {
		callable := &mockCallable{}
		tool, err := NewCallableTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		require.NoError(t, err)

		ch, err := tool.StreamExecute(context.Background(), &ExecutionRequest{})
		require.NoError(t, err)

		res := <-ch
		assert.Equal(t, "mock", res)
	})

	t.Run("StreamExecute_Streaming", func(t *testing.T) {
		callable := &mockStreamingCallable{}
		tool, err := NewCallableTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		require.NoError(t, err)

		ch, err := tool.StreamExecute(context.Background(), &ExecutionRequest{})
		require.NoError(t, err)

		res := <-ch
		assert.Equal(t, "streaming-mock", res)
	})
}
