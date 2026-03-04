// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestA2ABridgeMiddleware_Execute(t *testing.T) {
	ctx := context.Background()
	manager := NewRecursiveContextManager()
	bridge := NewA2ABridgeMiddleware(manager)

	// Test 1: Not a tools/call request
	t.Run("Not tools/call", func(t *testing.T) {
		req := &mcp.ListToolsRequest{}
		called := false
		next := func(ctx context.Context, method string, r mcp.Request) (mcp.Result, error) {
			called = true
			return &mcp.ListToolsResult{}, nil
		}

		res, err := bridge.Execute(ctx, "tools/list", req, next)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.True(t, called)
	})

	// Test 2: Not an mcp.CallToolRequest
	t.Run("Not mcp.CallToolRequest", func(t *testing.T) {
		req := &mcp.ListToolsRequest{}
		called := false
		next := func(ctx context.Context, method string, r mcp.Request) (mcp.Result, error) {
			called = true
			return &mcp.CallToolResult{}, nil
		}

		res, err := bridge.Execute(ctx, "tools/call", req, next)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.True(t, called)
	})

	// Test 3: Standard tool call
	t.Run("Standard tool call", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{"arg1": "val1"})
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{
				Name:      "standard_tool",
				Arguments: args,
			},
		}
		called := false
		next := func(ctx context.Context, method string, r mcp.Request) (mcp.Result, error) {
			called = true
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "standard response",
					},
				},
			}, nil
		}

		res, err := bridge.Execute(ctx, "tools/call", req, next)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.True(t, called)
		callRes := res.(*mcp.CallToolResult)
		textContent := callRes.Content[0].(*mcp.TextContent)
		assert.Equal(t, "standard response", textContent.Text)
	})

	// Test 4: A2A agent call
	t.Run("A2A agent call", func(t *testing.T) {
		args, _ := json.Marshal(map[string]interface{}{"task": "Write a Python script"})
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{
				Name:      "call_agent_coding_specialist",
				Arguments: args,
			},
		}
		called := false
		next := func(ctx context.Context, method string, r mcp.Request) (mcp.Result, error) {
			called = true
			return nil, nil
		}

		res, err := bridge.Execute(ctx, "tools/call", req, next)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.False(t, called)

		callRes := res.(*mcp.CallToolResult)
		assert.Len(t, callRes.Content, 1)
		textContent := callRes.Content[0].(*mcp.TextContent)
		assert.Contains(t, textContent.Text, "A2A Bridge: Successfully forwarded task to coding_specialist. Session ID:")

		// Extract session ID to verify it was stored
		var sessionID string
		for id := range manager.sessions {
			sessionID = id
		}
		assert.NotEmpty(t, sessionID)
		assert.Contains(t, textContent.Text, sessionID)
	})

	// Test 5: A2A agent call with string arguments
	t.Run("A2A agent call with string args", func(t *testing.T) {
		args, _ := json.Marshal("just a string")
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{
				Name:      "call_agent_research",
				Arguments: args,
			},
		}
		called := false
		next := func(ctx context.Context, method string, r mcp.Request) (mcp.Result, error) {
			called = true
			return nil, nil
		}

		res, err := bridge.Execute(ctx, "tools/call", req, next)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.False(t, called)

		callRes := res.(*mcp.CallToolResult)
		assert.Len(t, callRes.Content, 1)
		textContent := callRes.Content[0].(*mcp.TextContent)
		assert.Contains(t, textContent.Text, "A2A Bridge: Successfully forwarded task to research. Session ID:")
	})

	// Test 6: A2A agent call with nil arguments
	t.Run("A2A agent call with nil args", func(t *testing.T) {
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{
				Name: "call_agent_empty",
			},
		}
		called := false
		next := func(ctx context.Context, method string, r mcp.Request) (mcp.Result, error) {
			called = true
			return nil, nil
		}

		res, err := bridge.Execute(ctx, "tools/call", req, next)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.False(t, called)

		callRes := res.(*mcp.CallToolResult)
		assert.Len(t, callRes.Content, 1)
		textContent := callRes.Content[0].(*mcp.TextContent)
		assert.Contains(t, textContent.Text, "empty")
	})

	// Test 7: nil CallToolRequest
	t.Run("nil CallToolRequest", func(t *testing.T) {
		var req *mcp.CallToolRequest
		called := false
		next := func(ctx context.Context, method string, r mcp.Request) (mcp.Result, error) {
			called = true
			return &mcp.CallToolResult{}, nil
		}

		res, err := bridge.Execute(ctx, "tools/call", req, next)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.True(t, called)
	})
}
