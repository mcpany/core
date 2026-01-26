// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestCallableTool(t *testing.T) {
	t.Parallel()
	// Setup
	toolDef := configv1.ToolDefinition_builder{Name: proto.String("test-tool")}.Build()
	serviceConfig := configv1.UpstreamServiceConfig_builder{Id: proto.String("test-service")}.Build()

	mockCallable := &MockCallable{}

	ct, err := NewCallableTool(toolDef, serviceConfig, mockCallable, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, ct)

	assert.Equal(t, mockCallable, ct.Callable())
	assert.NotNil(t, ct.Tool())
	assert.Equal(t, "test-tool", ct.Tool().GetName())
	assert.Nil(t, ct.GetCacheConfig())
}

type MockCallable struct {
	CallFunc func(context.Context, *ExecutionRequest) (any, error)
}

func (m *MockCallable) Call(ctx context.Context, req *ExecutionRequest) (any, error) {
	if m.CallFunc != nil {
		return m.CallFunc(ctx, req)
	}
	return "result", nil
}

func (m *MockCallable) Parameters() map[string]any {
	return nil
}

func TestCallableTool_Execute(t *testing.T) {
	t.Parallel()
	mockC := &MockCallable{}
	ct, _ := NewCallableTool(
		configv1.ToolDefinition_builder{Name: proto.String("t")}.Build(),
		configv1.UpstreamServiceConfig_builder{Id: proto.String("s")}.Build(),
		mockC, nil, nil)

	res, err := ct.Execute(context.Background(), &ExecutionRequest{})
	assert.NoError(t, err)
	assert.Equal(t, "result", res)
}
