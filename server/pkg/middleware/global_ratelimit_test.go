// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestGlobalRateLimitMiddleware_Allow(t *testing.T) {
	// Setup config: 10 RPS, Burst 10
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             10,
		Storage:           configv1.RateLimitConfig_STORAGE_MEMORY,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_IP,
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg)

	// Mock next handler
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	// Mock context with IP
	ctx := util.ContextWithRemoteIP(context.Background(), "127.0.0.1")

	// Should pass
	for i := 0; i < 5; i++ {
		res, err := mw.Execute(ctx, "tools/call", nil, next)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	}
}

func TestGlobalRateLimitMiddleware_Block(t *testing.T) {
	// Setup config: 1 RPS, Burst 1
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 1,
		Burst:             1,
		Storage:           configv1.RateLimitConfig_STORAGE_MEMORY,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_IP,
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg)

	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	ctx := util.ContextWithRemoteIP(context.Background(), "127.0.0.1")

	// First call succeeds
	res, err := mw.Execute(ctx, "tools/call", nil, next)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	// Second call fails (burst is 1)
	res, err = mw.Execute(ctx, "tools/call", nil, next)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
	assert.Nil(t, res)
}

func TestGlobalRateLimitMiddleware_Disabled(t *testing.T) {
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         false,
		RequestsPerSecond: 1,
		Burst:             1,
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg)

	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	ctx := context.Background()

	// Should pass multiple times even if logic would block
	for i := 0; i < 10; i++ {
		res, err := mw.Execute(ctx, "tools/call", nil, next)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	}
}

func TestGlobalRateLimitMiddleware_KeyByUserID(t *testing.T) {
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 1,
		Burst:             1,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_USER_ID,
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg)
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	// User A
	ctxA := auth.ContextWithUser(context.Background(), "userA")
	// User B
	ctxB := auth.ContextWithUser(context.Background(), "userB")

	// User A Call 1 (OK)
	_, err := mw.Execute(ctxA, "tools/call", nil, next)
	assert.NoError(t, err)

	// User B Call 1 (OK) - separate bucket
	_, err = mw.Execute(ctxB, "tools/call", nil, next)
	assert.NoError(t, err)

	// User A Call 2 (Blocked)
	_, err = mw.Execute(ctxA, "tools/call", nil, next)
	assert.Error(t, err)
}

func TestGlobalRateLimitMiddleware_KeyByGlobal(t *testing.T) {
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 1,
		Burst:             1,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_GLOBAL,
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg)
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	// Request 1 (OK)
	_, err := mw.Execute(context.Background(), "tools/call", nil, next)
	assert.NoError(t, err)

	// Request 2 (Blocked) - shared bucket regardless of context
	_, err = mw.Execute(context.Background(), "tools/call", nil, next)
	assert.Error(t, err)
}
