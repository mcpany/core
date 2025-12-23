// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"testing"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
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
			name:     "String",
			input:    "hello",
			expected: 5,
		},
		{
			name:     "Byte Slice",
			input:    []byte("hello"),
			expected: 5,
		},
		{
			name: "CallToolResult with Text",
			input: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "hello"},
				},
			},
			expected: 5,
		},
		{
			name: "CallToolResult with Image",
			input: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.ImageContent{Data: []byte("image")},
				},
			},
			expected: 5,
		},
		{
			name: "CallToolResult with EmbeddedResource",
			input: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.EmbeddedResource{
						Resource: &mcp.ResourceContents{
							Blob: []byte("blob"),
						},
					},
				},
			},
			expected: 4,
		},
		{
			name: "CallToolResult Mixed",
			input: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "a"},
					&mcp.ImageContent{Data: []byte("b")},
				},
			},
			expected: 2,
		},
		{
			name: "Default JSON",
			input: map[string]string{"key": "val"},
			// {"key":"val"} -> 13 bytes
			expected: 13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, middleware.CalculateOutputSize(tt.input))
		})
	}
}
