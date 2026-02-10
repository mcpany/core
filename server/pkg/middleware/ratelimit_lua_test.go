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
	"google.golang.org/protobuf/proto"
)

func TestRedisLimiter_RateZero_Bypass(t *testing.T) {
	// Setup miniredis
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer s.Close()

	// Connect Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	defer rdb.Close()

	// Inject Redis creator
	middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return redis.NewClient(opts) // Use real client connected to miniredis
	})
	defer middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return redis.NewClient(opts)
	})

	// Inject Time
	mockTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	middleware.SetTimeNowForTests(func() time.Time {
		return mockTime
	})
	defer middleware.SetTimeNowForTests(time.Now)

	// Configure Limiter: Rate 0, Burst 5
	// Note: We need to use the miniredis address in config
	config := configv1.RateLimitConfig_builder{
		Redis:             busproto.RedisBus_builder{Address: proto.String(s.Addr())}.Build(),
		RequestsPerSecond: 0,
		Burst:             5,
	}.Build()

	// Need to pass config to NewRedisLimiter
	// But NewRedisLimiter takes *configv1.RateLimitConfig
	// Wait, is it *configv1.RateLimitConfig?
	// Yes, NewRedisLimiter(serviceID, config)

	limiter, err := middleware.NewRedisLimiter("test_service", config)
	assert.NoError(t, err)
	defer limiter.Close()

	ctx := context.Background()

	// 1. Consume Burst (5 requests)
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx)
		assert.NoError(t, err)
		if !allowed {
			t.Fatalf("Request %d should be allowed", i)
		}
	}

	// 2. Verify blocked
	allowed, err := limiter.Allow(ctx)
	assert.NoError(t, err)
	if allowed {
		t.Fatalf("Should be blocked after burst")
	}

	// 3. Advance time and keep spamming
	// With Rate=0, TTL is set to 60s (hardcoded in script).
	// We want to prove that if we spam for > 60s, the key expires if not refreshed.
	// If refreshed, it stays alive.

	step := 2 * time.Second
	duration := 70 * time.Second
	totalPassed := time.Duration(0)

	for totalPassed < duration {
		// Advance Redis time (affects TTL)
		s.FastForward(step)
		// Advance App time (affects token calculation)
		mockTime = mockTime.Add(step)
		totalPassed += step

		// Make a request (blocked)
		// This SHOULD refresh the TTL if implemented correctly.
		allowed, err := limiter.Allow(ctx)
		assert.NoError(t, err)

		// If bug exists, at around 60s, the key expires, and we get a fresh burst.
		if allowed {
			t.Fatalf("Bypass detected! Request allowed at T+%v (should be blocked forever with Rate=0)", totalPassed)
		}
	}
}
