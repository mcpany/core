// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-redis/redismock/v9"
	busproto "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestGlobalRateLimitMiddleware_Redis(t *testing.T) {
	db, mockRedis := redismock.NewClientMock()
	SetRedisClientCreatorForTests(func(_ *redis.Options) *redis.Client {
		return db
	})
	defer SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return redis.NewClient(opts)
	})

	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             10,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_GLOBAL,
		Redis: busproto.RedisBus_builder{
			Address: proto.String("localhost:6379"),
		}.Build(),
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg)

	// Test UpdateConfig
	cfg2 := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 20,
		Burst:             20,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_GLOBAL,
		Redis: busproto.RedisBus_builder{
			Address: proto.String("localhost:6379"),
		}.Build(),
	}.Build()
	mw.UpdateConfig(cfg2)
	assert.Equal(t, float64(20), mw.config.GetRequestsPerSecond())

	// Test Redis Execute
	s := redis.NewScript(RedisRateLimitScript)
	mockRedis.ExpectEvalSha(s.Hash(), []string{"ratelimit:global:global"}, 20.0, 20, 1).SetVal(int64(1))

	res, err := mw.Execute(context.Background(), "test", nil, func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
}

func TestGlobalRateLimitMiddleware_GetPartitionKey_APIKey(t *testing.T) {
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             10,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_API_KEY,
	}.Build()
	mw := NewGlobalRateLimitMiddleware(cfg)

	// 1. In Context
	ctx := auth.ContextWithAPIKey(context.Background(), "my-key")
	pk := mw.getPartitionKey(ctx, cfg.GetKeyBy())
	assert.Contains(t, pk, "apikey:")

	// 2. HTTP Request X-API-Key
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "http-key")
	ctx = context.WithValue(context.Background(), HTTPRequestContextKey, req)
	pk = mw.getPartitionKey(ctx, cfg.GetKeyBy())
	assert.Contains(t, pk, "apikey:")

	// 3. HTTP Request Authorization
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer token")
	ctx = context.WithValue(context.Background(), HTTPRequestContextKey, req)
	pk = mw.getPartitionKey(ctx, cfg.GetKeyBy())
	assert.Contains(t, pk, "auth:")

	// 4. None
	pk = mw.getPartitionKey(context.Background(), cfg.GetKeyBy())
	assert.Equal(t, "apikey:none", pk)
}

func TestGlobalRateLimitMiddleware_GetPartitionKey_Default(t *testing.T) {
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled: true,
	}.Build()
	mw := NewGlobalRateLimitMiddleware(cfg)
	// Some unknown config value
	pk := mw.getPartitionKey(context.Background(), configv1.RateLimitConfig_KeyBy(999))
	assert.Equal(t, "global", pk)
}

func TestGlobalRateLimitMiddleware_RedisConfigHash(t *testing.T) {
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             10,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_GLOBAL,
		Redis: busproto.RedisBus_builder{
			Address: proto.String("localhost:6379"),
			Password: proto.String("pass"),
			Db: proto.Int32(1),
		}.Build(),
	}.Build()
	mw := NewGlobalRateLimitMiddleware(cfg)

	hash1 := mw.calculateConfigHash(cfg.GetRedis())
	assert.NotEmpty(t, hash1)

	cfg2 := configv1.RateLimitConfig_builder{
		Redis: busproto.RedisBus_builder{
			Address: proto.String("localhost:6380"),
		}.Build(),
	}.Build()
	hash2 := mw.calculateConfigHash(cfg2.GetRedis())
	assert.NotEqual(t, hash1, hash2)
}

func TestGlobalRateLimitMiddleware_RedisErrorPassesThrough(t *testing.T) {
	db, mockRedis := redismock.NewClientMock()
	SetRedisClientCreatorForTests(func(_ *redis.Options) *redis.Client {
		return db
	})
	defer SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return redis.NewClient(opts)
	})

	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             10,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_GLOBAL,
		Redis: busproto.RedisBus_builder{
			Address: proto.String("localhost:6379"),
		}.Build(),
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg)

	// Simulate error when executing Allow
	s := redis.NewScript(RedisRateLimitScript)
	mockRedis.ExpectEvalSha(s.Hash(), []string{"ratelimit:global:global"}, 10.0, 10, 1).SetErr(context.DeadlineExceeded)

	// Fail open on error
	res, err := mw.Execute(context.Background(), "test", nil, func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestGlobalRateLimitMiddleware_MissingRedisConfig(t *testing.T) {
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             10,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_GLOBAL,
		// Missing Redis block
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg)

	// Since Redis config is missing, getLimiter fails, but we fail open
	res, err := mw.Execute(context.Background(), "test", nil, func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestGlobalRateLimitMiddleware_RedisClientCaching(t *testing.T) {
	db, mockRedis := redismock.NewClientMock()
	SetRedisClientCreatorForTests(func(_ *redis.Options) *redis.Client {
		return db
	})
	defer SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return redis.NewClient(opts)
	})

	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             10,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_GLOBAL,
		Redis: busproto.RedisBus_builder{
			Address: proto.String("localhost:6379"),
		}.Build(),
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg)

	// Trigger getRedisClient to create it and cache it
	s := redis.NewScript(RedisRateLimitScript)
	mockRedis.ExpectEvalSha(s.Hash(), []string{"ratelimit:global:global"}, 10.0, 10, 1).SetVal(int64(1))
	_, err := mw.Execute(context.Background(), "test", nil, func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	})
	assert.NoError(t, err)

	// Execute again, should load from cache
	mockRedis.ExpectEvalSha(s.Hash(), []string{"ratelimit:global:global"}, 10.0, 10, 1).SetVal(int64(1))
	_, err = mw.Execute(context.Background(), "test", nil, func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	})
	assert.NoError(t, err)

	assert.NoError(t, mockRedis.ExpectationsWereMet())
}

func TestGlobalRateLimitMiddleware_RedisConfigUpdateForcesNewLimiter(t *testing.T) {
	db, mockRedis := redismock.NewClientMock()
	SetRedisClientCreatorForTests(func(_ *redis.Options) *redis.Client {
		return db
	})
	defer SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return redis.NewClient(opts)
	})

	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             10,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_GLOBAL,
		Redis: busproto.RedisBus_builder{
			Address: proto.String("localhost:6379"),
		}.Build(),
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg)

	// Trigger getLimiter, this creates the redis limiter
	s := redis.NewScript(RedisRateLimitScript)
	mockRedis.ExpectEvalSha(s.Hash(), []string{"ratelimit:global:global"}, 10.0, 10, 1).SetVal(int64(1))
	_, err := mw.Execute(context.Background(), "test", nil, func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	})
	assert.NoError(t, err)

	// Now update the config with a different Redis DB or Password to change the hash
	cfg2 := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 20,
		Burst:             20,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_GLOBAL,
		Redis: busproto.RedisBus_builder{
			Address: proto.String("localhost:6379"),
			Db: proto.Int32(2),
		}.Build(),
	}.Build()

	mw.UpdateConfig(cfg2)

	// Execute again. The cached limiter will have a different config hash,
	// so it should create a new limiter.
	mockRedis.ExpectEvalSha(s.Hash(), []string{"ratelimit:global:global"}, 20.0, 20, 1).SetVal(int64(1))
	_, err = mw.Execute(context.Background(), "test", nil, func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	})
	assert.NoError(t, err)

	assert.NoError(t, mockRedis.ExpectationsWereMet())
}

func TestGlobalRateLimitMiddleware_TypeMismatch(t *testing.T) {
	cfg1 := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             10,
		Storage:           configv1.RateLimitConfig_STORAGE_MEMORY,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_GLOBAL,
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg1)

	// First execute populates the cache with LocalLimiter
	_, err := mw.Execute(context.Background(), "test", nil, func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	})
	assert.NoError(t, err)

	// Update to Redis storage
	db, mockRedis := redismock.NewClientMock()
	SetRedisClientCreatorForTests(func(_ *redis.Options) *redis.Client {
		return db
	})
	defer SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return redis.NewClient(opts)
	})

	cfg2 := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             10,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_GLOBAL,
		Redis: busproto.RedisBus_builder{
			Address: proto.String("localhost:6379"),
		}.Build(),
	}.Build()

	mw.UpdateConfig(cfg2)

	// Execute again. Type mismatch (Local vs Redis) should force creating a new limiter
	s := redis.NewScript(RedisRateLimitScript)
	mockRedis.ExpectEvalSha(s.Hash(), []string{"ratelimit:global:global"}, 10.0, 10, 1).SetVal(int64(1))
	_, err = mw.Execute(context.Background(), "test", nil, func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	})
	assert.NoError(t, err)

	assert.NoError(t, mockRedis.ExpectationsWereMet())
}

func TestGlobalRateLimitMiddleware_GetRedisClient_Concurrent(t *testing.T) {
	// Need to test LoadOrStore race condition path
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             10,
		Storage:           configv1.RateLimitConfig_STORAGE_REDIS,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_GLOBAL,
		Redis: busproto.RedisBus_builder{
			Address: proto.String("localhost:6379"),
		}.Build(),
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg)
	// getRedisClient is private but we can reach it
	// Create multiple goroutines to trigger LoadOrStore loading logic
	done := make(chan struct{})

	// Create 100 goroutines to call getRedisClient simultaneously
	for i := 0; i < 100; i++ {
		go func() {
			mw.getRedisClient(cfg.GetRedis())
			done <- struct{}{}
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
	// Test should not crash and coverage should be higher for getRedisClient
}

func TestGlobalRateLimitMiddleware_GetPartitionKey_IP(t *testing.T) {
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled: true,
		KeyBy:     configv1.RateLimitConfig_KEY_BY_IP,
	}.Build()
	mw := NewGlobalRateLimitMiddleware(cfg)

	// Valid IP context
	ctxWithIP := util.ContextWithRemoteIP(context.Background(), "192.168.1.1")
	assert.Equal(t, "ip:192.168.1.1", mw.getPartitionKey(ctxWithIP, cfg.GetKeyBy()))

	// Missing IP context
	assert.Equal(t, "ip:unknown", mw.getPartitionKey(context.Background(), cfg.GetKeyBy()))
}

func TestGlobalRateLimitMiddleware_GetPartitionKey_UserID(t *testing.T) {
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled: true,
		KeyBy:     configv1.RateLimitConfig_KEY_BY_USER_ID,
	}.Build()
	mw := NewGlobalRateLimitMiddleware(cfg)

	// Valid User ID context
	ctxWithUser := auth.ContextWithUser(context.Background(), "user-123")
	assert.Equal(t, "user:user-123", mw.getPartitionKey(ctxWithUser, cfg.GetKeyBy()))

	// Missing User ID context
	assert.Equal(t, "user:anonymous", mw.getPartitionKey(context.Background(), cfg.GetKeyBy()))
}

func TestGlobalRateLimitMiddleware_Burst0DefaultsTo1(t *testing.T) {
	cfg := configv1.RateLimitConfig_builder{
		IsEnabled:         true,
		RequestsPerSecond: 10,
		Burst:             0,
		Storage:           configv1.RateLimitConfig_STORAGE_MEMORY,
		KeyBy:             configv1.RateLimitConfig_KEY_BY_GLOBAL,
	}.Build()

	mw := NewGlobalRateLimitMiddleware(cfg)

	limiter, err := mw.getLimiter(context.Background(), cfg)
	assert.NoError(t, err)
	// Underlying limiter should have burst=1
	assert.NotNil(t, limiter)

	// Ensure we can allow 1 request
	allowed, err := limiter.Allow(context.Background())
	assert.NoError(t, err)
	assert.True(t, allowed)
}
