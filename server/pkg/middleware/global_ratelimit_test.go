// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"
	"testing"

	"github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
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

func TestGlobalRateLimitMiddleware_UpdateConfig(t *testing.T) {
	cfg1 := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 1,
		Burst:             1,
		Storage:           configv1.RateLimitConfig_STORAGE_MEMORY,
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg1)
	assert.Equal(t, cfg1, mw.config)

	cfg2 := configv1.RateLimitConfig_builder{
		IsEnabled:         false,
		RequestsPerSecond: 10,
		Burst:             5,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
	}.Build()

	mw.UpdateConfig(cfg2)
	assert.Equal(t, cfg2, mw.config)
}

func TestGlobalRateLimitMiddleware_CalculateConfigHash(t *testing.T) {
	mw := NewGlobalRateLimitMiddleware(nil)

	bus1 := bus.RedisBus_builder{
		Address:  proto.String("localhost:6379"),
		Password: proto.String("pass"),
		Db:       proto.Int32(0),
	}.Build()

	bus2 := bus.RedisBus_builder{
		Address:  proto.String("localhost:6379"),
		Password: proto.String("pass"),
		Db:       proto.Int32(0),
	}.Build()

	bus3 := bus.RedisBus_builder{
		Address:  proto.String("localhost:6380"),
		Password: proto.String("pass"),
		Db:       proto.Int32(0),
	}.Build()

	hash1 := mw.calculateConfigHash(bus1)
	hash2 := mw.calculateConfigHash(bus2)
	hash3 := mw.calculateConfigHash(bus3)

	assert.Equal(t, hash1, hash2, "identical configs should produce identical hashes")
	assert.NotEqual(t, hash1, hash3, "different configs should produce different hashes")
}

func TestGlobalRateLimitMiddleware_GetPartitionKey(t *testing.T) {
	mw := NewGlobalRateLimitMiddleware(nil)

	tests := []struct {
		name     string
		keyBy    configv1.RateLimitConfig_KeyBy
		ctxSetup func() context.Context
		expected string
	}{
		{
			name:  "IP with IP context",
			keyBy: configv1.RateLimitConfig_KEY_BY_IP,
			ctxSetup: func() context.Context {
				return util.ContextWithRemoteIP(context.Background(), "192.168.1.1")
			},
			expected: "ip:192.168.1.1",
		},
		{
			name:  "IP without IP context",
			keyBy: configv1.RateLimitConfig_KEY_BY_IP,
			ctxSetup: func() context.Context {
				return context.Background()
			},
			expected: "ip:unknown",
		},
		{
			name:  "User ID with user context",
			keyBy: configv1.RateLimitConfig_KEY_BY_USER_ID,
			ctxSetup: func() context.Context {
				return auth.ContextWithUser(context.Background(), "user123")
			},
			expected: "user:user123",
		},
		{
			name:  "User ID without user context",
			keyBy: configv1.RateLimitConfig_KEY_BY_USER_ID,
			ctxSetup: func() context.Context {
				return context.Background()
			},
			expected: "user:anonymous",
		},
		{
			name:  "API Key with API key context",
			keyBy: configv1.RateLimitConfig_KEY_BY_API_KEY,
			ctxSetup: func() context.Context {
				return auth.ContextWithAPIKey(context.Background(), "secret-key")
			},
			expected: hashKey("apikey:", "secret-key"),
		},
		{
			name:  "API Key with X-API-Key HTTP header",
			keyBy: configv1.RateLimitConfig_KEY_BY_API_KEY,
			ctxSetup: func() context.Context {
				req, _ := http.NewRequest("GET", "/", nil)
				req.Header.Set("X-API-Key", "header-secret-key")
				return context.WithValue(context.Background(), HTTPRequestContextKey, req)
			},
			expected: hashKey("apikey:", "header-secret-key"),
		},
		{
			name:  "API Key with Authorization HTTP header",
			keyBy: configv1.RateLimitConfig_KEY_BY_API_KEY,
			ctxSetup: func() context.Context {
				req, _ := http.NewRequest("GET", "/", nil)
				req.Header.Set("Authorization", "Bearer token123")
				return context.WithValue(context.Background(), HTTPRequestContextKey, req)
			},
			expected: hashKey("auth:", "Bearer token123"),
		},
		{
			name:  "API Key without API key context or headers",
			keyBy: configv1.RateLimitConfig_KEY_BY_API_KEY,
			ctxSetup: func() context.Context {
				return context.Background()
			},
			expected: "apikey:none",
		},
		{
			name:  "Global key",
			keyBy: configv1.RateLimitConfig_KEY_BY_GLOBAL,
			ctxSetup: func() context.Context {
				return context.Background()
			},
			expected: "global",
		},
		{
			name:  "Unspecified key type",
			keyBy: configv1.RateLimitConfig_KEY_BY_UNSPECIFIED,
			ctxSetup: func() context.Context {
				return context.Background()
			},
			expected: "global",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.ctxSetup()
			key := mw.getPartitionKey(ctx, tc.keyBy)
			assert.Equal(t, tc.expected, key)
		})
	}
}

func TestGlobalRateLimitMiddleware_GetRedisClient(t *testing.T) {
	mw := NewGlobalRateLimitMiddleware(nil)

	bus1 := bus.RedisBus_builder{
		Address:  proto.String("localhost:6379"),
		Password: proto.String("pass"),
		Db:       proto.Int32(0),
	}.Build()

	bus2 := bus.RedisBus_builder{
		Address:  proto.String("localhost:6380"),
		Password: proto.String("pass"),
		Db:       proto.Int32(1),
	}.Build()

	client1 := mw.getRedisClient(bus1)
	assert.NotNil(t, client1)

	client1Again := mw.getRedisClient(bus1)
	assert.True(t, client1 == client1Again, "should return cached client")

	client2 := mw.getRedisClient(bus2)
	assert.NotNil(t, client2)
	assert.True(t, client1 != client2, "should create new client for different config")
}

func TestGlobalRateLimitMiddleware_GetLimiter(t *testing.T) {
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             0, // Burst should be clamped to 1
		Storage:           configv1.RateLimitConfig_STORAGE_MEMORY,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_IP,
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg)
	ctx := util.ContextWithRemoteIP(context.Background(), "127.0.0.1")

	// Create a new limiter
	limiter1, err := mw.getLimiter(ctx, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, limiter1)

	// Fetch cached limiter
	limiter2, err := mw.getLimiter(ctx, cfg)
	assert.NoError(t, err)
	assert.Equal(t, limiter1, limiter2)

	// Update limiter config
	cfgUpdated := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 20,
		Burst:             5,
		Storage:           configv1.RateLimitConfig_STORAGE_MEMORY,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_IP,
	}.Build()

	limiter3, err := mw.getLimiter(ctx, cfgUpdated)
	assert.NoError(t, err)
	assert.Equal(t, limiter1, limiter3)

	// Switch to REDIS without redis config
	cfgRedisNoConfig := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 20,
		Burst:             5,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_IP,
	}.Build()

	limiter4, err := mw.getLimiter(ctx, cfgRedisNoConfig)
	assert.Error(t, err)
	assert.Nil(t, limiter4)
	assert.Contains(t, err.Error(), "redis config is missing")
}
