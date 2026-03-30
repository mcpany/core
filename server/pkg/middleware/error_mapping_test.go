// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestErrorMappingMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		inputErr    error
		expectedMsg string
	}{
		{
			name:        "No Error",
			inputErr:    nil,
			expectedMsg: "",
		},
		{
			name:        "Not Found Error",
			inputErr:    errors.New("open /missing/file.txt: no such file or directory"),
			expectedMsg: "Resource not found: open /missing/file.txt: no such file or directory",
		},
		{
			name:        "Permission Denied Error",
			inputErr:    fmt.Errorf("access denied: %v", "permission denied"),
			expectedMsg: "Permission denied: access denied: permission denied",
		},
		{
			name:        "Timeout Error",
			inputErr:    errors.New("deadline exceeded during request"),
			expectedMsg: "Upstream timeout: deadline exceeded during request",
		},
		{
			name:        "Generic Error",
			inputErr:    errors.New("something went terribly wrong"),
			expectedMsg: "Upstream execution error: something went terribly wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapped := NewErrorMappingMiddleware()

			req := &tool.ExecutionRequest{ToolName: "test_tool"}

			nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
				return "success", tt.inputErr
			}

			res, err := mapped.Execute(context.Background(), req, nextFunc)

			// ErrorMappingMiddleware should swallow the error and return it as a result
			assert.NoError(t, err)

			if tt.inputErr == nil {
				assert.Equal(t, "success", res)
			} else {
				// Must be of type *mcp.CallToolResult
				mcpRes, ok := res.(*mcp.CallToolResult)
				assert.True(t, ok, "Result must be of type *mcp.CallToolResult")
				assert.True(t, mcpRes.IsError)
				assert.Len(t, mcpRes.Content, 1)
				textContent, ok := mcpRes.Content[0].(*mcp.TextContent)
				assert.True(t, ok)
				assert.Contains(t, textContent.Text, tt.expectedMsg)
			}
		})
	}
}
