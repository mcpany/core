// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestGlobalRateLimitMiddleware_UpdateConfig(t *testing.T) {
	// Test that updating configuration dynamically works
	cfg := &configv1.RateLimitConfig{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             10,
		Storage:           configv1.RateLimitConfig_STORAGE_MEMORY,
	}

	mw := NewGlobalRateLimitMiddleware(cfg)
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}
	ctx := context.Background()

	// Initial check
	_, err := mw.Execute(ctx, "method", nil, next)
	assert.NoError(t, err)

	// Update to disable
	newCfg := &configv1.RateLimitConfig{
		IsEnabled: false,
	}
	mw.UpdateConfig(newCfg)

	// Should still work (disabled)
	_, err = mw.Execute(ctx, "method", nil, next)
	assert.NoError(t, err)
}

func TestGlobalRateLimitMiddleware_KeyByAPIKey(t *testing.T) {
	cfg := &configv1.RateLimitConfig{
		IsEnabled:         true,
		RequestsPerSecond: 1,
		Burst:             1,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_API_KEY,
	}
	mw := NewGlobalRateLimitMiddleware(cfg)
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	// 1. Context with API Key
	ctx1 := auth.ContextWithAPIKey(context.Background(), "apikey1")
	_, err := mw.Execute(ctx1, "method", nil, next)
	assert.NoError(t, err)

	// 2. HTTP Request with X-API-Key
	req2, _ := http.NewRequest("POST", "/", nil)
	req2.Header.Set("X-API-Key", "apikey2")
	ctx2 := context.WithValue(context.Background(), HTTPRequestContextKey, req2)
	_, err = mw.Execute(ctx2, "method", nil, next)
	assert.NoError(t, err)

	// 3. HTTP Request with Authorization
	req3, _ := http.NewRequest("POST", "/", nil)
	req3.Header.Set("Authorization", "Bearer token3")
	ctx3 := context.WithValue(context.Background(), HTTPRequestContextKey, req3)
	_, err = mw.Execute(ctx3, "method", nil, next)
	assert.NoError(t, err)

	// 4. No API Key (fallback)
	ctx4 := context.Background()
	_, err = mw.Execute(ctx4, "method", nil, next)
	assert.NoError(t, err)
}

func TestGlobalRateLimitMiddleware_Redis(t *testing.T) {
	// Setup mock redis
	db, mock := redismock.NewClientMock()

	// Inject mock creator
	originalCreator := redisClientCreator
	defer func() { redisClientCreator = originalCreator }()
	SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return db
	})

	// Setup Time
	originalTime := timeNow
	defer func() { timeNow = originalTime }()
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	SetTimeNowForTests(func() time.Time { return now })

	rBus := &bus.RedisBus{}
	rBus.SetAddress("localhost:6379")
	cfg := &configv1.RateLimitConfig{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             20,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		Redis:             rBus,
	}
	mw := NewGlobalRateLimitMiddleware(cfg)
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}
	ctx := context.Background()

	// Expect Redis script call
	// Keys: ratelimit:global:global
	// Args: rps, burst, now(micro), cost(1)

	key := "ratelimit:global:global"

	// First call - allowed
	mock.ExpectEvalSha(redis.NewScript(RedisRateLimitScript).Hash(), []string{key},
		10.0, 20, now.UnixMicro(), 1).
		SetVal(int64(1))

	_, err := mw.Execute(ctx, "method", nil, next)
	assert.NoError(t, err)

	// Second call - blocked
	mock.ExpectEvalSha(redis.NewScript(RedisRateLimitScript).Hash(), []string{key},
		10.0, 20, now.UnixMicro(), 1).
		SetVal(int64(0))

	_, err = mw.Execute(ctx, "method", nil, next)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "global rate limit exceeded")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGlobalRateLimitMiddleware_Redis_MissingConfig(t *testing.T) {
	cfg := &configv1.RateLimitConfig{
		IsEnabled: true,
		Storage:   configv1.RateLimitConfig_STORAGE_REDIS,
		Redis:     nil, // Missing Redis config
	}
	mw := NewGlobalRateLimitMiddleware(cfg)
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}
	ctx := context.Background()

	// Should fail open (allow)
	_, err := mw.Execute(ctx, "method", nil, next)
	assert.NoError(t, err)
}

func TestGlobalRateLimitMiddleware_Redis_ClientCache(t *testing.T) {
	db, _ := redismock.NewClientMock()

	// Inject mock creator
	originalCreator := redisClientCreator
	defer func() { redisClientCreator = originalCreator }()

	createCount := 0
	SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		createCount++
		return db
	})

	rBus := &bus.RedisBus{}
	rBus.SetAddress("localhost:6379")
	cfg := &configv1.RateLimitConfig{
		IsEnabled:         true,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		Redis:             rBus,
	}
	mw := NewGlobalRateLimitMiddleware(cfg)
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}
	ctx := context.Background()

	// First call - creates client
	mw.Execute(ctx, "method", nil, next)
	assert.Equal(t, 1, createCount)

	// Second call - reuses client
	mw.Execute(ctx, "method", nil, next)
	assert.Equal(t, 1, createCount)

	// Change config
	rBus2 := &bus.RedisBus{}
	rBus2.SetAddress("localhost:6380") // Different address
	cfg2 := &configv1.RateLimitConfig{
		IsEnabled:         true,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		Redis:             rBus2,
	}
	// We need to trigger getLimiter again with new config.
	// But Execute uses m.config. So we need to update it.
	mw.UpdateConfig(cfg2)

	// Third call - should create new client because config hash changed
	mw.Execute(ctx, "method", nil, next)
	assert.Equal(t, 2, createCount)
}
