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

type mockCallable struct{}

func (m *mockCallable) Call(ctx context.Context, req *ExecutionRequest) (any, error) {
	return nil, nil
}

func TestBaseTool_New(t *testing.T) {
	t.Parallel()

	toolDef := &configv1.ToolDefinition{
		Name:        "test-tool",
		Description: "A test tool",
	}

	serviceConfig := &configv1.UpstreamServiceConfig{}
	callable := &mockCallable{}

	// Test valid creation
	bt, err := newBaseTool(toolDef, serviceConfig, callable, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, bt)
	assert.Equal(t, "test-tool", bt.Tool().GetName())
	assert.Equal(t, "A test tool", bt.Tool().GetDescription())

	// Test invalid creation (e.g., missing name triggers conversion error)
	invalidDef := &configv1.ToolDefinition{}
	_, err = newBaseTool(invalidDef, serviceConfig, callable, nil, nil)
	assert.Error(t, err)
}

func TestBaseTool_MCPTool(t *testing.T) {
	t.Parallel()

	toolDef := &configv1.ToolDefinition{
		Name:        "test-mcp-tool",
		Description: "A test mcp tool",
	}

	bt, err := newBaseTool(toolDef, nil, nil, nil, nil)
	require.NoError(t, err)

	mcpT := bt.MCPTool()
	require.NotNil(t, mcpT)
	assert.Equal(t, "test-mcp-tool", mcpT.Name)
	assert.Equal(t, "A test mcp tool", mcpT.Description)

	// Test caching/once behavior
	mcpT2 := bt.MCPTool()
	assert.Same(t, mcpT, mcpT2)
}

func TestBaseTool_Methods(t *testing.T) {
	t.Parallel()

	toolDef := &configv1.ToolDefinition{Name: "t"}
	bt, err := newBaseTool(toolDef, nil, nil, nil, nil)
	require.NoError(t, err)

	assert.False(t, bt.IsStreaming())
	assert.Nil(t, bt.GetCacheConfig())

	ch, err := bt.StreamExecute(context.Background(), &ExecutionRequest{})
	assert.Nil(t, ch)
	assert.NoError(t, err)
}
