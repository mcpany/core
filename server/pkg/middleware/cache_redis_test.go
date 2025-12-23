// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestCachingMiddleware_RedisStorage(t *testing.T) {
	// Setup Redis Mock
	originalCreator := middleware.GetRedisClientCreatorForTests()
	defer middleware.SetRedisClientCreatorForTests(originalCreator)

	// Mock Redis Client
	// Since we cannot easily mock the redis.Client struct completely (it's a struct, not interface),
	// we use a miniredis or just rely on the fact that we can intercept creation.
	// But `redis.NewClient` returns `*redis.Client`.
	// For this unit test, verifying that it TRIES to create a redis client is enough.
	// We can't actually connect to Redis unless we spin one up.

	// Strategy: Mock `redisClientCreator` to return a client connected to an invalid address,
	// but we assert that it was called with correct options.
	// Or even better: use `github.com/alicebob/miniredis/v2` if available.
	// Let's check availability. If not, we just check configuration logic.

	var capturedOpts *redis.Options
	middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		capturedOpts = opts
		return redis.NewClient(opts)
	})

	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)

	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("redis-tool"),
			ServiceId: proto.String("redis-service"),
		}.Build(),
		cacheConfig: configv1.CacheConfig_builder{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(1 * time.Hour),
			Storage:   configv1.CacheConfig_STORAGE_REDIS.Enum(),
			Redis: bus.RedisBus_builder{
				Address:  proto.String("localhost:6379"),
				Password: proto.String("secret"),
				Db:       proto.Int32(0),
			}.Build(),
		}.Build(),
	}

	req := &tool.ExecutionRequest{
		ToolName: "redis-service.redis-tool",
	}
	ctx := tool.NewContextWithTool(context.Background(), testTool)
	nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return successResult, nil
	}

	// Execute. It will try to create Redis client.
	// It might fail to connect inside `Execute` when calling `cache.Get`, but `getCacheManager` call happens first.
	// We just want to verify `getCacheManager` logic triggers redis creation.
	// However, `getCacheManager` creates the store. The store might try to ping?
	// `redis_store.NewRedis` doesn't ping immediately.
	// `cache.Get` will try to use the client.

	// Since we don't have a real Redis, execution might fail or hang.
	// But we can verify that `capturedOpts` is set.
	// We expect `Execute` to fail or log error if redis is unreachable, but we want to confirm CONFIG flow.

	// To make this safe, we can mock `redisClientCreator` to return nil (if handled) or check if we can verify before `Execute`?
	// `Execute` calls `getCacheManager`.
	// We can invoke `Execute` and expect it to proceed.
	// If redis fails, `Execute` handles error (fail open)?
	// My implementation: `cacheManager.Get` returns error -> MISS -> execute tool -> `cacheManager.Set` (error log).
	// So it should "work" (return result) even if Redis is down.

	res, err := cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res)

	// Verify Redis options were passed correctly
	require.NotNil(t, capturedOpts)
	assert.Equal(t, "localhost:6379", capturedOpts.Addr)
	assert.Equal(t, "secret", capturedOpts.Password)
	assert.Equal(t, 0, capturedOpts.DB)
}
