// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCalculateOutputSize(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int
	}{
		{
			name:     "Nil",
			input:    nil,
			expected: 0,
		},
		{
			name: "CallToolResult with TextContent",
			input: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "hello"},
				},
			},
			expected: 5,
		},
		{
			name: "CallToolResult with ImageContent",
			input: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.ImageContent{Data: []byte("image_data")},
				},
			},
			expected: 10,
		},
		{
			name: "CallToolResult with EmbeddedResource",
			input: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.EmbeddedResource{
						Resource: &mcp.ResourceContents{
							Blob: []byte("resource_blob"),
						},
					},
				},
			},
			expected: 13,
		},
		{
			name: "CallToolResult with Mixed Content",
			input: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "ab"},
					&mcp.ImageContent{Data: []byte("cd")},
				},
			},
			expected: 4,
		},
		{
			name:     "String",
			input:    "hello world",
			expected: 11,
		},
		{
			name:     "Bytes",
			input:    []byte("byte slice"),
			expected: 10,
		},
		{
			name:     "Map (JSON fallback)",
			input:    map[string]string{"k": "v"},
			expected: 9, // {"k":"v"}
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, calculateOutputSize(tt.input))
		})
	}
}

func TestCachingMiddleware_Clear(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock not needed for NewCachingMiddleware call but needed for args type
	mockToolManager := tool.NewMockManagerInterface(ctrl)

	m := NewCachingMiddleware(mockToolManager)

	// Set something in cache
	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName:   "test_tool",
		ToolInputs: []byte("{}"),
	}
	key := m.getCacheKey(req)
	err := m.cache.Set(ctx, key, "value")
	require.NoError(t, err)

	// Verify it's there
	val, err := m.cache.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, "value", val)

	// Clear
	err = m.Clear(ctx)
	require.NoError(t, err)

	// Verify it's gone
	_, err = m.cache.Get(ctx, key)
	assert.Error(t, err)
}
