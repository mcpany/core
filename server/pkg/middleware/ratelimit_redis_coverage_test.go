package middleware

import (
	"testing"

	"github.com/go-redis/redismock/v9"
	bus "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestRedisLimiter_GetConfigHash(t *testing.T) {
	db, _ := redismock.NewClientMock()

	rps := 10.0
	burst := int64(20)

	address := "127.0.0.1:6379"
	password := "pass"
	dbIdx := int32(0)

	// Using setters to avoid direct struct initialization issues with hidden fields in opaque mode
	redisBus := &bus.RedisBus{}
	redisBus.SetAddress(address)
	redisBus.SetPassword(password)
	redisBus.SetDb(dbIdx)

	config := configv1.RateLimitConfig_builder{
		RequestsPerSecond: rps,
		Burst:             burst,
		Redis:             redisBus,
	}.Build()

	limiter := NewRedisLimiterWithClient(db, "service", "", "partition", config)

	expectedHash := "127.0.0.1:6379|pass|0"
	assert.Equal(t, expectedHash, limiter.GetConfigHash())

	// Test with no redis config
	limiterNoRedis := NewRedisLimiterWithClient(db, "service", "", "partition", configv1.RateLimitConfig_builder{}.Build())
	assert.Equal(t, "", limiterNoRedis.GetConfigHash())
}
