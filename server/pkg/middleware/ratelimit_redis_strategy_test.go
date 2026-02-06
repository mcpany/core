// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"testing"

	"github.com/go-redis/redismock/v9"
	busproto "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestRedisStrategy(t *testing.T) {
	// Setup mock creator
	var createdClients []*redis.Client
	middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		db, _ := redismock.NewClientMock()
		createdClients = append(createdClients, db)
		return db
	})
	defer middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return redis.NewClient(opts)
	})

	t.Run("Create and Caching", func(t *testing.T) {
		// Reset created clients
		createdClients = nil
		strategy := middleware.NewRedisStrategy()
		ctx := context.Background()

		config1 := configv1.RateLimitConfig_builder{
			Redis: busproto.RedisBus_builder{
				Address:  proto.String("localhost:6379"),
				Password: proto.String("pass"),
				Db:       proto.Int32(0),
			}.Build(),
			RequestsPerSecond: 10,
			Burst:             10,
		}.Build()

		// 1. First creation
		limiter1, err := strategy.Create(ctx, "serviceA", "scope", "partition", config1)
		require.NoError(t, err)
		assert.NotNil(t, limiter1)
		assert.Len(t, createdClients, 1)
		client1 := createdClients[0]

		// 2. Second creation with SAME config -> Should reuse cached client
		limiter2, err := strategy.Create(ctx, "serviceA", "scope", "partition", config1)
		require.NoError(t, err)
		assert.NotNil(t, limiter2)
		assert.Len(t, createdClients, 1, "Should reuse existing client")

		// Verify underlying client is the same (indirectly via createdClients list)
		// We can't easily access the client inside limiter interface, but we know Create returns a wrapper around it.
		// If createdClients len is 1, it means it didn't call the creator again.

		// 3. Third creation with DIFFERENT config -> Should create new client
		config2 := configv1.RateLimitConfig_builder{
			Redis: busproto.RedisBus_builder{
				Address:  proto.String("localhost:6379"),
				Password: proto.String("pass"),
				Db:       proto.Int32(1), // Different DB
			}.Build(),
			RequestsPerSecond: 10,
			Burst:             10,
		}.Build()

		limiter3, err := strategy.Create(ctx, "serviceA", "scope", "partition", config2)
		require.NoError(t, err)
		assert.NotNil(t, limiter3)
		assert.Len(t, createdClients, 2, "Should create new client for different config")
		client2 := createdClients[1]
		assert.NotEqual(t, client1, client2)

		// 4. Different Service ID -> Should create new client even if config is same as first one (Wait, really?)
		// The code uses serviceID as the primary key.
		// if val, ok := s.redisClients.Load(serviceID); ok ...
		// So if serviceID is different, it will NOT find it and create a new one.
		// Even if the Redis config is identical?
		// Yes, the implementation keys by serviceID.
		limiter4, err := strategy.Create(ctx, "serviceB", "scope", "partition", config1)
		require.NoError(t, err)
		assert.NotNil(t, limiter4)
		assert.Len(t, createdClients, 3, "Should create new client for different serviceID")
	})

	t.Run("Missing Redis Config", func(t *testing.T) {
		strategy := middleware.NewRedisStrategy()
		ctx := context.Background()
		config := configv1.RateLimitConfig_builder{
			// No Redis config
			RequestsPerSecond: 10,
		}.Build()

		_, err := strategy.Create(ctx, "serviceC", "", "", config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis config is missing")
	})

	t.Run("ShouldUpdate", func(t *testing.T) {
		strategy := middleware.NewRedisStrategy()
		ctx := context.Background()

		config1 := configv1.RateLimitConfig_builder{
			Redis: busproto.RedisBus_builder{
				Address:  proto.String("localhost:6379"),
				Password: proto.String("pass"),
				Db:       proto.Int32(0),
			}.Build(),
			RequestsPerSecond: 10,
			Burst:             10,
		}.Build()

		// Create a limiter
		limiter, err := strategy.Create(ctx, "serviceUpdate", "", "", config1)
		require.NoError(t, err)

		// 1. Same config -> ShouldUpdate = false
		assert.False(t, strategy.ShouldUpdate(limiter, config1))

		// 2. Different Redis config -> ShouldUpdate = true
		config2 := configv1.RateLimitConfig_builder{
			Redis: busproto.RedisBus_builder{
				Address:  proto.String("new-host"),
				Password: proto.String("pass"),
				Db:       proto.Int32(0),
			}.Build(),
			RequestsPerSecond: 10,
			Burst:             10,
		}.Build()
		assert.True(t, strategy.ShouldUpdate(limiter, config2))

		// 3. Missing Redis config -> ShouldUpdate = true
		config3 := configv1.RateLimitConfig_builder{
			RequestsPerSecond: 10,
		}.Build()
		assert.True(t, strategy.ShouldUpdate(limiter, config3))

		// 4. Non-Redis limiter (mock) passed in -> ShouldUpdate = true
		// (Simulate if we switched from Local to Redis)
		// We use NewLocalStrategy to create a valid LocalLimiter (which is a Limiter but not RedisLimiter)
		localLimiter, _ := middleware.NewLocalStrategy().Create(ctx, "foo", "", "", config1)
		assert.True(t, strategy.ShouldUpdate(localLimiter, config1))
	})
}
