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

func TestRedisLimiter_Integration_ZeroRateBypass(t *testing.T) {
	// 1. Setup Miniredis
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	// 2. Setup Redis Client
	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	defer rdb.Close()

	// Inject the client creator
	middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return rdb
	})
	defer middleware.SetRedisClientCreatorForTests(nil)

	// 3. Mock Time
	currentTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	middleware.SetTimeNowForTests(func() time.Time {
		return currentTime
	})
	defer middleware.SetTimeNowForTests(nil)

	// 4. Create Limiter with Rate=0, Burst=1
	config := configv1.RateLimitConfig_builder{
		Redis:             busproto.RedisBus_builder{Address: proto.String(s.Addr())}.Build(),
		RequestsPerSecond: 0,
		Burst:             1,
	}.Build()

	l, err := middleware.NewRedisLimiter("zero-rate-service", config)
	require.NoError(t, err)
	defer l.Close()

	ctx := context.Background()

	// 5. First request: Should be allowed (Burst=1)
	allowed, err := l.Allow(ctx)
	require.NoError(t, err)
	assert.True(t, allowed, "First request should be allowed")
	t.Logf("T=0: Allowed=%v", allowed)

	// Check key TTL
	ttl := s.TTL("ratelimit:zero-rate-service")
	t.Logf("T=0: Key TTL=%v", ttl)

	// 6. Second request: Should be denied
	allowed, err = l.Allow(ctx)
	require.NoError(t, err)
	assert.False(t, allowed, "Second request should be denied")
	t.Logf("T=0: Allowed=%v (Expected False)", allowed)

	// 7. Advance time past the default TTL (60s)
	for i := 0; i < 65; i++ {
		currentTime = currentTime.Add(1 * time.Second)
		s.FastForward(1 * time.Second)

		// Every 10 seconds, check status
		if i%10 == 0 {
			ttl := s.TTL("ratelimit:zero-rate-service")
			t.Logf("T=%d: Key TTL=%v", i, ttl)
		}

		// Try to access
		res, err := l.Allow(ctx)
		require.NoError(t, err)
		if res {
			t.Logf("BUG DETECTED at T=%d: Request allowed (should be denied)", i)
			t.Fail()
		}
	}

	// 8. Probe again at T=65
	allowed, err = l.Allow(ctx)
	require.NoError(t, err)
	t.Logf("T=65: Allowed=%v", allowed)

	// If bug exists, this should be TRUE because key expired and reset.
	// If fix exists (or behavior is different), it should be FALSE.
	// We Assert FALSE to catch the bug (if it returns TRUE, test fails).
	assert.False(t, allowed, "Should be denied after TTL if we kept probing")
}
