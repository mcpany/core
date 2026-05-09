// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestESBMiddleware_Disabled(t *testing.T) {
	middleware := NewESBMiddleware(&configv1.Middleware{Disabled: true})

	req := mcp.CallToolRequest{}
	nextCalled := false
	var next mcp.MethodHandler = func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.CallToolResult{}, nil
	}

	res, err := middleware.Execute(context.Background(), "callTool", &req, next)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, nextCalled)
}

func TestESBMiddleware_NotCallToolRequest(t *testing.T) {
	middleware := NewESBMiddleware(&configv1.Middleware{Disabled: false})

	// Use InitializeRequest, which should bypass the ESB logic
	req := mcp.InitializeRequest{}
	nextCalled := false
	var next mcp.MethodHandler = func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.InitializeResult{}, nil
	}

	res, err := middleware.Execute(context.Background(), "initialize", &req, next)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, nextCalled)
}

func TestESBMiddleware_MissingMissionIntent(t *testing.T) {
	middleware := NewESBMiddleware(&configv1.Middleware{Disabled: false})

	req := mcp.CallToolRequest{}
	nextCalled := false
	var next mcp.MethodHandler = func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.CallToolResult{}, nil
	}

	// Context is empty, so no mission intent
	ctx := context.Background()

	res, err := middleware.Execute(ctx, "callTool", &req, next)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.False(t, nextCalled)

	// Verify the result is an error message from ESB
	toolRes, ok := res.(*mcp.CallToolResult)
	assert.True(t, ok)
	assert.True(t, toolRes.IsError)
	assert.Len(t, toolRes.Content, 1)
	textContent, ok := toolRes.Content[0].(*mcp.TextContent)
	assert.True(t, ok)
	assert.Equal(t, "ESB Error: Missing x-mission-intent header/context.", textContent.Text)
}

func TestESBMiddleware_MissingEntanglementShard(t *testing.T) {
	middleware := NewESBMiddleware(&configv1.Middleware{Disabled: false})

	req := mcp.CallToolRequest{}
	nextCalled := false
	var next mcp.MethodHandler = func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.CallToolResult{}, nil
	}

	// Context has mission intent but no entanglement shard
	ctx := context.WithValue(context.Background(), missionIntentKey, "intent-123")

	res, err := middleware.Execute(ctx, "callTool", &req, next)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.False(t, nextCalled)

	// Verify the result is an error message from ESB
	toolRes, ok := res.(*mcp.CallToolResult)
	assert.True(t, ok)
	assert.True(t, toolRes.IsError)
	assert.Len(t, toolRes.Content, 1)
	textContent, ok := toolRes.Content[0].(*mcp.TextContent)
	assert.True(t, ok)
	assert.Equal(t, "ESB Error: Missing x-entanglement-shard header/context.", textContent.Text)
}

func TestESBMiddleware_Success_TypedContext(t *testing.T) {
	middleware := NewESBMiddleware(&configv1.Middleware{Disabled: false})

	req := mcp.CallToolRequest{}
	nextCalled := false
	var next mcp.MethodHandler = func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.CallToolResult{}, nil
	}

	// Context has both requirements as typed esbContextKey
	ctx := context.WithValue(context.Background(), missionIntentKey, "intent-123")
	ctx = context.WithValue(ctx, entanglementShardKey, "shard-456")

	res, err := middleware.Execute(ctx, "callTool", &req, next)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, nextCalled)
}

func TestESBMiddleware_Success_StringContext(t *testing.T) {
	middleware := NewESBMiddleware(&configv1.Middleware{Disabled: false})

	req := mcp.CallToolRequest{}
	nextCalled := false
	var next mcp.MethodHandler = func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.CallToolResult{}, nil
	}

	// Context has both requirements as regular strings
	ctx := context.WithValue(context.Background(), "x-mission-intent", "intent-123")
	ctx = context.WithValue(ctx, "x-entanglement-shard", "shard-456")

	res, err := middleware.Execute(ctx, "callTool", &req, next)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, nextCalled)
}
