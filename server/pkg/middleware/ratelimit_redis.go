package middleware

import (
	"context"
	"fmt"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/redis/go-redis/v9"
)

var redisClientCreator = redis.NewClient

// SetRedisClientCreatorForTests sets the redis client creator for tests.
func SetRedisClientCreatorForTests(creator func(opts *redis.Options) *redis.Client) {
	redisClientCreator = creator
}

var timeNow = time.Now

// SetTimeNowForTests sets the time.Now function for tests.
func SetTimeNowForTests(nowFunc func() time.Time) {
	timeNow = nowFunc
}

// RedisLimiter implements rate limiting using Redis.
type RedisLimiter struct {
	client *redis.Client
	key    string
	rps    float64
	burst  int
}

// NewRedisLimiter creates a new RedisLimiter.
func NewRedisLimiter(serviceID string, config *configv1.RateLimitConfig) (*RedisLimiter, error) {
	return NewRedisLimiterWithPartition(serviceID, "", config)
}

// NewRedisLimiterWithPartition creates a new RedisLimiter with a partition key.
func NewRedisLimiterWithPartition(serviceID, partitionKey string, config *configv1.RateLimitConfig) (*RedisLimiter, error) {
	if config.GetRedis() == nil {
		return nil, fmt.Errorf("redis config is missing")
	}

	redisConfig := config.GetRedis()
	opts := &redis.Options{
		Addr:     redisConfig.GetAddress(),
		Password: redisConfig.GetPassword(),
		DB:       int(redisConfig.GetDb()),
	}
	client := redisClientCreator(opts)

	key := fmt.Sprintf("ratelimit:%s", serviceID)
	if partitionKey != "" {
		key = fmt.Sprintf("%s:%s", key, partitionKey)
	}

	return &RedisLimiter{
		client: client,
		key:    key,
		rps:    config.GetRequestsPerSecond(),
		burst:  int(config.GetBurst()),
	}, nil
}

// NewRedisLimiterWithClient creates a new RedisLimiter with an existing client.
func NewRedisLimiterWithClient(client *redis.Client, serviceID, partitionKey string, config *configv1.RateLimitConfig) *RedisLimiter {
	key := fmt.Sprintf("ratelimit:%s", serviceID)
	if partitionKey != "" {
		key = fmt.Sprintf("%s:%s", key, partitionKey)
	}
	return &RedisLimiter{
		client: client,
		key:    key,
		rps:    config.GetRequestsPerSecond(),
		burst:  int(config.GetBurst()),
	}
}

// RedisRateLimitScript is the Lua script used for rate limiting.
const RedisRateLimitScript = `
    local key = KEYS[1]
    local rate = tonumber(ARGV[1])
    local burst = tonumber(ARGV[2])
    local now = tonumber(ARGV[3]) -- microseconds
    local cost = tonumber(ARGV[4])

    local fields = redis.call("HMGET", key, "tokens", "last_refill")
    local tokens = tonumber(fields[1])
    local last_refill = tonumber(fields[2])

    if not tokens then
        tokens = burst
        last_refill = now
    end

    local delta = (now - last_refill) / 1000000 -- seconds
    local filled_tokens = math.min(burst, tokens + (delta * rate))

    if filled_tokens >= cost then
        local new_tokens = filled_tokens - cost
        redis.call("HMSET", key, "tokens", new_tokens, "last_refill", now)

        -- Expire key after enough time to refill completely + buffer
        local ttl = 60
        if rate > 0 then
             ttl = math.ceil(burst / rate * 2)
        end
        if ttl < 1 then ttl = 1 end
        redis.call("EXPIRE", key, ttl)
        return 1
    end

    return 0
    `

// Allow checks if the request is allowed.
func (l *RedisLimiter) Allow(ctx context.Context) (bool, error) {
	now := timeNow().UnixMicro()

	// Use float64 for rate to handle fractional rates
	cmd := l.client.Eval(ctx, RedisRateLimitScript, []string{l.key}, l.rps, l.burst, now, 1)
	if cmd.Err() != nil {
		return false, cmd.Err()
	}

	res, ok := cmd.Val().(int64)
	if !ok {
		// Redis might return different type?
		return false, fmt.Errorf("unexpected return type from redis script: %T", cmd.Val())
	}

	return res == 1, nil
}

// Update updates the limiter configuration.
func (l *RedisLimiter) Update(rps float64, burst int) {
	l.rps = rps
	l.burst = burst
}

// Close closes the Redis client.
func (l *RedisLimiter) Close() error {
	return l.client.Close()
}
