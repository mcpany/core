// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestToolActivityFeed(t *testing.T) {
	b, _ := bus.NewProvider(nil)
	m := tool.NewManager(b)

	// Mock tool that succeeds
	mockHandler := func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "success"}, nil
	}
	t1 := configv1.ToolDefinition_builder{
		Name:      proto.String("success_tool"),
		ServiceId: proto.String("s1"),
	}.Build()

	// Create a callable tool using NewCallableTool.
	// But we need a concrete implementation.
	// We can't easily mock tool.Tool interface without implementing all methods.
	// Let's rely on AddTool logic. But AddTool expects a Tool interface.
	// `tool.NewCallableTool` returns a *CallableTool which implements Tool.
	// We need `tool.Callable` interface.
	// Let's define a simple callable.

	callable1 := &mockCallable{handler: mockHandler}
	ct1, err := tool.NewCallableTool(t1, nil, callable1, nil, nil)
	require.NoError(t, err)
	err = m.AddTool(ct1)
	require.NoError(t, err)

	// Mock tool that fails
	mockHandlerFail := func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
		return nil, fmt.Errorf("fail")
	}
	t2 := configv1.ToolDefinition_builder{
		Name:      proto.String("fail_tool"),
		ServiceId: proto.String("s1"),
	}.Build()
	callable2 := &mockCallable{handler: mockHandlerFail}
	ct2, err := tool.NewCallableTool(t2, nil, callable2, nil, nil)
	require.NoError(t, err)
	err = m.AddTool(ct2)
	require.NoError(t, err)

	// Execute success tool
	req1 := &tool.ExecutionRequest{
		ToolName: "s1.success_tool",
	}
	_, err = m.ExecuteTool(context.Background(), req1)
	require.NoError(t, err)

	// Execute fail tool
	req2 := &tool.ExecutionRequest{
		ToolName: "s1.fail_tool",
	}
	_, err = m.ExecuteTool(context.Background(), req2)
	require.Error(t, err)

	// Verify history
	history := m.GetHistory()
	require.Len(t, history, 2)

	assert.Equal(t, "s1.success_tool", history[0].ToolName)
	assert.True(t, history[0].Success)
	assert.Empty(t, history[0].Error)

	assert.Equal(t, "s1.fail_tool", history[1].ToolName)
	assert.False(t, history[1].Success)
	assert.Equal(t, "fail", history[1].Error)
}

type mockCallable struct {
	handler func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error)
}

func (c *mockCallable) Call(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return c.handler(ctx, req.Arguments)
}
