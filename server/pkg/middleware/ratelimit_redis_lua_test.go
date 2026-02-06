// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	busproto "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// setupMiniredis creates a miniredis instance and configures the RedisLimiter to use it.
// It returns the miniredis instance, a cleanup function, and the RedisLimiter.
func setupMiniredis(t *testing.T, rps float64, burst int) (*miniredis.Miniredis, func(), *middleware.RedisLimiter) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	// Override the client creator to use miniredis address
	middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
	})

	// Override time to be deterministic
	currentTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	middleware.SetTimeNowForTests(func() time.Time {
		return currentTime
	})

	config := configv1.RateLimitConfig_builder{
		Redis:             busproto.RedisBus_builder{Address: proto.String(s.Addr())}.Build(),
		RequestsPerSecond: rps,
		Burst:             int64(burst),
	}.Build()

	l, err := middleware.NewRedisLimiter("test_service", config)
	require.NoError(t, err)

	cleanup := func() {
		s.Close()
		// Restore original functions
		middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
			return redis.NewClient(opts)
		})
		middleware.SetTimeNowForTests(time.Now)
	}

	return s, cleanup, l
}

func TestRedisLimiter_Lua_Basic(t *testing.T) {
	// RPS: 10, Burst: 10
	_, cleanup, l := setupMiniredis(t, 10, 10)
	defer cleanup()

	// 1. Basic Allow
	allowed, err := l.Allow(context.Background())
	assert.NoError(t, err)
	assert.True(t, allowed, "First request should be allowed")
}

func TestRedisLimiter_Lua_Burst(t *testing.T) {
	// RPS: 1, Burst: 5
	_, cleanup, l := setupMiniredis(t, 1, 5)
	defer cleanup()

	// Consume all 5 tokens
	for i := 0; i < 5; i++ {
		allowed, err := l.Allow(context.Background())
		assert.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed (burst)", i+1)
	}

	// 6th request should fail (no time passed)
	allowed, err := l.Allow(context.Background())
	assert.NoError(t, err)
	assert.False(t, allowed, "Request 6 should be denied (burst exceeded)")
}

func TestRedisLimiter_Lua_Refill(t *testing.T) {
	// RPS: 10, Burst: 10
	// 1 token refills every 100ms
	_, cleanup, l := setupMiniredis(t, 10, 10)
	defer cleanup()

	// Setup time control
	startTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	currentTime := startTime
	middleware.SetTimeNowForTests(func() time.Time {
		return currentTime
	})

	// Consume all 10 tokens
	for i := 0; i < 10; i++ {
		allowed, err := l.Allow(context.Background())
		require.True(t, allowed)
		require.NoError(t, err)
	}

	// Next should fail
	allowed, err := l.Allow(context.Background())
	assert.False(t, allowed)
	assert.NoError(t, err)

	// Advance time by 500ms (should refill 5 tokens: 10 RPS * 0.5s = 5)
	currentTime = startTime.Add(500 * time.Millisecond)

	// Should be able to consume 5 tokens
	for i := 0; i < 5; i++ {
		allowed, err := l.Allow(context.Background())
		assert.True(t, allowed, "Refilled request %d should be allowed", i+1)
		assert.NoError(t, err)
	}

	// 6th should fail
	allowed, err = l.Allow(context.Background())
	assert.False(t, allowed, "Should be empty again")
	assert.NoError(t, err)
}

func TestRedisLimiter_Lua_Cost(t *testing.T) {
	// RPS: 10, Burst: 10
	_, cleanup, l := setupMiniredis(t, 10, 10)
	defer cleanup()

	// Consume 5 tokens at once
	allowed, err := l.AllowN(context.Background(), 5)
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Consume remaining 5
	allowed, err = l.AllowN(context.Background(), 5)
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Try to consume 1 more
	allowed, err = l.Allow(context.Background())
	assert.False(t, allowed)
}

func TestRedisLimiter_Lua_CostExceedsBurst(t *testing.T) {
	// RPS: 10, Burst: 10
	_, cleanup, l := setupMiniredis(t, 10, 10)
	defer cleanup()

	// Try to consume 11 tokens (more than burst)
	// The script should handle this: min(burst, tokens) - cost
	// tokens=10. filled=10. 10 >= 11 is False.
	allowed, err := l.AllowN(context.Background(), 11)
	assert.NoError(t, err)
	assert.False(t, allowed, "Should not allow request costing more than burst")
}

func TestRedisLimiter_Lua_Expiry(t *testing.T) {
	// RPS: 1, Burst: 10
	// TTL calculation in script:
	// ttl = ceil(burst / rate * 2) = ceil(10 / 1 * 2) = 20 seconds
	s, cleanup, l := setupMiniredis(t, 1, 10)
	defer cleanup()

	// Make a request to create the key
	_, err := l.Allow(context.Background())
	require.NoError(t, err)

	// Check TTL in Redis
	// Key is "ratelimit:test_service"
	ttl := s.TTL("ratelimit:test_service")
	assert.True(t, ttl > 0, "Key should have a TTL")
	assert.True(t, ttl <= 20*time.Second, "TTL should be around 20s")
}
