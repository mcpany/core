package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryLimiter(t *testing.T) {
	limiter := NewInMemoryLimiter(10, 1)

	// Allow the first request
	assert.True(t, limiter.Allow(), "First request should be allowed")

	// Exhaust the burst capacity
	for i := 0; i < 10; i++ {
		limiter.Allow()
	}

	// The next request should be denied
	assert.False(t, limiter.Allow(), "Request after exhausting burst should be denied")

	// Wait for the limiter to refill
	time.Sleep(100 * time.Millisecond)

	// The next request should be allowed
	assert.True(t, limiter.Allow(), "Request after refill should be allowed")
}
