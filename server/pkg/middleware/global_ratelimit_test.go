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

func newRateLimitConfig(enabled bool, rps float64, burst int64, storage *configv1.RateLimitConfig_Storage, keyBy *configv1.RateLimitConfig_KeyBy) *configv1.RateLimitConfig {
	c := &configv1.RateLimitConfig{}
	c.SetIsEnabled(enabled)
	c.SetRequestsPerSecond(rps)
	c.SetBurst(burst)
	if storage != nil {
		c.SetStorage(*storage)
	}
	if keyBy != nil {
		c.SetKeyBy(*keyBy)
	}
	return c
}

func TestGlobalRateLimitMiddleware_Allow(t *testing.T) {
	// Setup config: 10 RPS, Burst 10
	cfg := newRateLimitConfig(true, 10, 10, configv1.RateLimitConfig_STORAGE_MEMORY.Enum(), configv1.RateLimitConfig_KEY_BY_IP.Enum())

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
	cfg := newRateLimitConfig(true, 1, 1, configv1.RateLimitConfig_STORAGE_MEMORY.Enum(), configv1.RateLimitConfig_KEY_BY_IP.Enum())

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
	cfg := newRateLimitConfig(false, 1, 1, nil, nil)

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
	cfg := newRateLimitConfig(true, 1, 1, nil, configv1.RateLimitConfig_KEY_BY_USER_ID.Enum())

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
	cfg := newRateLimitConfig(true, 1, 1, nil, configv1.RateLimitConfig_KEY_BY_GLOBAL.Enum())

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
