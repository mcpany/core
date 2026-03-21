// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	router := NewRouter()

	t.Run("register and get handler", func(t *testing.T) {
		expectedResult := &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "test result"}},
		}
		handler := func(_ context.Context, _ mcp.Request) (mcp.Result, error) {
			return expectedResult, nil
		}
		router.Register("test/method", handler)

		retrievedHandler, ok := router.GetHandler("test/method")
		assert.True(t, ok, "handler should be found")
		assert.NotNil(t, retrievedHandler, "handler should not be nil")

		res, err := retrievedHandler(context.Background(), nil)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, res)
	})

	t.Run("get non-existent handler", func(t *testing.T) {
		_, ok := router.GetHandler("non/existent")
		assert.False(t, ok, "handler should not be found")
	})
}
