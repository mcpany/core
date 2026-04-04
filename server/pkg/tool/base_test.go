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

type mockCallable struct{}

func (m *mockCallable) Call(ctx context.Context, req *ExecutionRequest) (any, error) {
	return "mock", nil
}

func TestBaseTool_New(t *testing.T) {
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
		"properties": map[string]interface{}{
			"param1": map[string]interface{}{
				"type": "string",
			},
		},
	})

	outputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
	})

	callable := &mockCallable{}

	t.Run("newBaseTool", func(t *testing.T) {
		base, err := newBaseTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		require.NoError(t, err)
		assert.NotNil(t, base)
		assert.Equal(t, callable, base.callable)
		assert.Equal(t, serviceConfig, base.serviceConfig)
		assert.NotNil(t, base.tool)
		assert.Equal(t, "test_tool", base.tool.GetName())
	})

	t.Run("Tool", func(t *testing.T) {
		base, err := newBaseTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		require.NoError(t, err)

		pbTool := base.Tool()
		assert.NotNil(t, pbTool)
		assert.Equal(t, "test_tool", pbTool.GetName())
	})

	t.Run("MCPTool", func(t *testing.T) {
		base, err := newBaseTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		require.NoError(t, err)

		mcpTool := base.MCPTool()
		assert.NotNil(t, mcpTool)
		assert.Equal(t, "test_tool", mcpTool.Name)
		assert.Equal(t, "A test tool", mcpTool.Description)

		// Calling again should return the cached instance
		mcpTool2 := base.MCPTool()
		assert.Equal(t, mcpTool, mcpTool2)
	})

	t.Run("IsStreaming", func(t *testing.T) {
		base, err := newBaseTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		require.NoError(t, err)
		assert.False(t, base.IsStreaming())
	})

	t.Run("StreamExecute", func(t *testing.T) {
		base, err := newBaseTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		require.NoError(t, err)

		ch, err := base.StreamExecute(context.Background(), &ExecutionRequest{})
		assert.Nil(t, ch)
		assert.NoError(t, err)
	})

	t.Run("GetCacheConfig", func(t *testing.T) {
		base, err := newBaseTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		require.NoError(t, err)
		assert.Nil(t, base.GetCacheConfig())
	})
}
