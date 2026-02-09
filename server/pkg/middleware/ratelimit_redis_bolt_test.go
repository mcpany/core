package middleware

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	busproto "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestRedisLimiter_RateZeroBypass(t *testing.T) {
	// Setup Miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Use our miniredis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	// Create limiter manually
	// Limiter Config: Rate=0, Burst=2
	redisConfig := busproto.RedisBus_builder{
		Address: proto.String(mr.Addr()),
	}.Build()

	config := configv1.RateLimitConfig_builder{
		RequestsPerSecond: 0, // Rate=0
		Burst:             2, // Burst=2
		Redis:             redisConfig,
	}.Build()

	limiter := NewRedisLimiterWithClient(client, "test-service", "", "", config)

	// Modify the global script to force TTL=1 for reproduction
	originalScript := redisRateLimitScript
	defer func() { redisRateLimitScript = originalScript }()

	// Replace "local ttl = 60" with "local ttl = 1" to enable fast expiry in test
	modifiedScriptSrc := strings.Replace(RedisRateLimitScript, "local ttl = 60", "local ttl = 1", 1)
	redisRateLimitScript = redis.NewScript(modifiedScriptSrc)

	// Mock time
	currentTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	originalTimeNow := timeNow
	defer func() { timeNow = originalTimeNow }()
	SetTimeNowForTests(func() time.Time {
		return currentTime
	})

	ctx := context.Background()

	// Step 1: Consume Burst (2 tokens)
	// T=0
	allowed, err := limiter.Allow(ctx)
	require.NoError(t, err)
	assert.True(t, allowed, "First request should be allowed")

	allowed, err = limiter.Allow(ctx)
	require.NoError(t, err)
	assert.True(t, allowed, "Second request should be allowed")

	// Burst exhausted.
	allowed, err = limiter.Allow(ctx)
	require.NoError(t, err)
	assert.False(t, allowed, "Third request should be blocked")

	// Step 2: Wait 0.5s. Try again. Should be blocked.
	// T=0.5
	currentTime = currentTime.Add(500 * time.Millisecond)
	mr.FastForward(500 * time.Millisecond)

	allowed, err = limiter.Allow(ctx)
	require.NoError(t, err)
	assert.False(t, allowed, "Should still be blocked after 0.5s")

	// Step 3: Wait another 0.6s. Total wait 1.1s.
	// Key expires at T=1.0s (TTL=1).
	// T=1.1
	currentTime = currentTime.Add(600 * time.Millisecond)
	mr.FastForward(600 * time.Millisecond)

	allowed, err = limiter.Allow(ctx)
	require.NoError(t, err)

	// ⚡ Bolt Verification: The bug should be fixed now.
	// We expect to be BLOCKED (false) because the key should have been refreshed even while blocked.
	assert.False(t, allowed, "Request should be blocked. If allowed, the key expired and bucket reset (Bug Persists).")
}
