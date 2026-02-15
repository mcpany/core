package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestLocalLimiter_Allow(t *testing.T) {
	// 10 requests per second, burst 1
	rl := rate.NewLimiter(rate.Limit(10), 1)
	l := &LocalLimiter{Limiter: rl}
	ctx := context.Background()

	// First request should be allowed
	allowed, err := l.Allow(ctx)
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Second request immediately after (burst is 1) might be allowed depending on timing,
	// but with burst 1 and high rate, it refills quickly.
	// To test blocking, we can use a very slow rate.
}

func TestLocalLimiter_AllowN(t *testing.T) {
	// 1 request per second, burst 5
	rl := rate.NewLimiter(rate.Limit(1), 5)
	l := &LocalLimiter{Limiter: rl}
	ctx := context.Background()

	// Request cost 2
	allowed, err := l.AllowN(ctx, 2)
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Tokens remaining: ~3.
	// Request cost 4 (exceeds remaining)
	allowed, err = l.AllowN(ctx, 4)
	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestLocalLimiter_Update(t *testing.T) {
	// Start with 1 RPS, Burst 1
	rl := rate.NewLimiter(rate.Limit(1), 1)
	l := &LocalLimiter{Limiter: rl}

	assert.Equal(t, rate.Limit(1), l.Limit())
	assert.Equal(t, 1, l.Burst())

	// Update to 10 RPS, Burst 5
	l.Update(10.0, 5)

	assert.Equal(t, rate.Limit(10.0), l.Limit())
	assert.Equal(t, 5, l.Burst())
}
