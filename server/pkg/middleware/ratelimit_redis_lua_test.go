// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"strconv"
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

// TestRedisLimiter_LuaScript verifies the correctness of the Redis Lua script logic
// using a miniredis instance which supports script execution.
func TestRedisLimiter_LuaScript(t *testing.T) {
	// Setup miniredis
	s := miniredis.RunT(t)

	// Inject miniredis client creator
	middleware.SetRedisClientCreatorForTests(func(_ *redis.Options) *redis.Client {
		return redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
	})
	defer middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return redis.NewClient(opts)
	})

	// Inject controlled time
	var currentTime time.Time
	middleware.SetTimeNowForTests(func() time.Time {
		return currentTime
	})
	defer middleware.SetTimeNowForTests(time.Now)

	// Helper to create config
	createConfig := func(rps float64, burst int64) *configv1.RateLimitConfig {
		return configv1.RateLimitConfig_builder{
			Redis:             busproto.RedisBus_builder{Address: proto.String(s.Addr())}.Build(),
			RequestsPerSecond: rps,
			Burst:             burst,
		}.Build()
	}

	tests := []struct {
		name          string
		rps           float64
		burst         int64
		actions       func(t *testing.T, l *middleware.RedisLimiter, s *miniredis.Miniredis)
	}{
		{
			name:  "Initial burst capacity",
			rps:   10,
			burst: 10,
			actions: func(t *testing.T, l *middleware.RedisLimiter, s *miniredis.Miniredis) {
				// Should allow 10 requests immediately
				for i := 0; i < 10; i++ {
					allowed, err := l.Allow(context.Background())
					require.NoError(t, err)
					assert.True(t, allowed, "Request %d should be allowed", i+1)
				}
				// 11th request should be denied
				allowed, err := l.Allow(context.Background())
				require.NoError(t, err)
				assert.False(t, allowed, "Request 11 should be denied")
			},
		},
		{
			name:  "Refill logic",
			rps:   1, // 1 request per second
			burst: 5,
			actions: func(t *testing.T, l *middleware.RedisLimiter, s *miniredis.Miniredis) {
				// Consume all 5 tokens
				allowed, err := l.AllowN(context.Background(), 5)
				require.NoError(t, err)
				assert.True(t, allowed)

				// Verify empty
				allowed, err = l.Allow(context.Background())
				require.NoError(t, err)
				assert.False(t, allowed)

				// Advance time by 1 second (1 token refilled)
				currentTime = currentTime.Add(1 * time.Second)

				// Should allow 1 request
				allowed, err = l.Allow(context.Background())
				require.NoError(t, err)
				assert.True(t, allowed)

				// Should be empty again
				allowed, err = l.Allow(context.Background())
				require.NoError(t, err)
				assert.False(t, allowed)
			},
		},
		{
			name:  "Fractional rate (0.5 RPS)",
			rps:   0.5, // 1 request every 2 seconds
			burst: 2,
			actions: func(t *testing.T, l *middleware.RedisLimiter, s *miniredis.Miniredis) {
				// Consume burst
				allowed, err := l.AllowN(context.Background(), 2)
				require.NoError(t, err)
				assert.True(t, allowed)

				// Advance 1 second (0.5 tokens refilled -> not enough for 1 request)
				currentTime = currentTime.Add(1 * time.Second)
				allowed, err = l.Allow(context.Background())
				require.NoError(t, err)
				assert.False(t, allowed, "Should not allow request with only 0.5 tokens")

				// Advance another 1 second (total 2 seconds -> 1 token refilled)
				currentTime = currentTime.Add(1 * time.Second)
				allowed, err = l.Allow(context.Background())
				require.NoError(t, err)
				assert.True(t, allowed, "Should allow request after 2 seconds")
			},
		},
		{
			name:  "Burst cap",
			rps:   10,
			burst: 5,
			actions: func(t *testing.T, l *middleware.RedisLimiter, s *miniredis.Miniredis) {
				// Advance time by a lot (e.g., 1 hour)
				currentTime = currentTime.Add(1 * time.Hour)

				// Even with huge time delta, should only have 'burst' tokens
				// Consume 5 tokens
				allowed, err := l.AllowN(context.Background(), 5)
				require.NoError(t, err)
				assert.True(t, allowed)

				// Next should fail
				allowed, err = l.Allow(context.Background())
				require.NoError(t, err)
				assert.False(t, allowed, "Should not exceed burst capacity")
			},
		},
		{
			name:  "Cost > 1",
			rps:   10,
			burst: 10,
			actions: func(t *testing.T, l *middleware.RedisLimiter, s *miniredis.Miniredis) {
				// Request cost 6
				allowed, err := l.AllowN(context.Background(), 6)
				require.NoError(t, err)
				assert.True(t, allowed)

				// Remaining tokens: 4. Request cost 5 should fail
				allowed, err = l.AllowN(context.Background(), 5)
				require.NoError(t, err)
				assert.False(t, allowed)

				// Request cost 4 should succeed
				allowed, err = l.AllowN(context.Background(), 4)
				require.NoError(t, err)
				assert.True(t, allowed)
			},
		},
		{
			name:  "Cost > Burst",
			rps:   10,
			burst: 5,
			actions: func(t *testing.T, l *middleware.RedisLimiter, s *miniredis.Miniredis) {
				// Request cost 6 when burst is 5 should always fail
				allowed, err := l.AllowN(context.Background(), 6)
				require.NoError(t, err)
				assert.False(t, allowed)
			},
		},
		{
			name:  "TTL Expiration",
			rps:   1,
			burst: 10,
			actions: func(t *testing.T, l *middleware.RedisLimiter, s *miniredis.Miniredis) {
				// Trigger initial write
				l.Allow(context.Background())

				// Calculate expected TTL from Lua script logic:
				// ttl = math.ceil(burst / rate * 2)
				// ttl = ceil(10 / 1 * 2) = 20 seconds

				// Miniredis stores keys. We can check TTL.
				// Key format: ratelimit:{serviceID}
				key := "ratelimit:test-service"

				ttl := s.TTL(key)
				assert.Greater(t, ttl, 0*time.Second)
				// The actual TTL might be slightly different due to implementation details but should be close to 20s
				assert.LessOrEqual(t, ttl, 20*time.Second)
				assert.GreaterOrEqual(t, ttl, 10*time.Second)

				// Advance time past TTL (miniredis clock)
				s.FastForward(21 * time.Second)

				// Key should be gone
				exists := s.Exists(key)
				assert.False(t, exists, "Key should verify expired")
			},
		},
		{
			name:  "Verify Token State in Redis",
			rps:   10,
			burst: 10,
			actions: func(t *testing.T, l *middleware.RedisLimiter, s *miniredis.Miniredis) {
				l.AllowN(context.Background(), 3)

				key := "ratelimit:test-service"
				// Verify tokens field
				val := s.HGet(key, "tokens")
				require.NotEmpty(t, val)

				tokens, err := strconv.ParseFloat(val, 64)
				require.NoError(t, err)

				// Should have consumed 3, so 7 remaining
				// Floating point comparison
				assert.InDelta(t, 7.0, tokens, 0.001)

				// Verify last_refill
				val = s.HGet(key, "last_refill")
				require.NotEmpty(t, val)
				lastRefill, err := strconv.ParseInt(val, 10, 64)
				require.NoError(t, err)
				assert.Equal(t, currentTime.UnixMicro(), lastRefill)
			},
		},
		{
			name: "High Rate High Burst",
			rps: 1000,
			burst: 1000,
			actions: func(t *testing.T, l *middleware.RedisLimiter, s *miniredis.Miniredis) {
				// Should allow 1000 requests
				allowed, err := l.AllowN(context.Background(), 1000)
				require.NoError(t, err)
				assert.True(t, allowed)

				// Next one fails
				allowed, err = l.Allow(context.Background())
				require.NoError(t, err)
				assert.False(t, allowed)

				// Advance 1ms (1 request refilled)
				currentTime = currentTime.Add(1 * time.Millisecond)
				allowed, err = l.Allow(context.Background())
				require.NoError(t, err)
				assert.True(t, allowed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.FlushAll()
			currentTime = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

			config := createConfig(tt.rps, tt.burst)
			l, err := middleware.NewRedisLimiter("test-service", config)
			require.NoError(t, err)

			tt.actions(t, l, s)

			l.Close()
		})
	}
}
