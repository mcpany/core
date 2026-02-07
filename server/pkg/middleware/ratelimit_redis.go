// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/redis/go-redis/v9"
)

var redisClientCreator = redis.NewClient

// SetRedisClientCreatorForTests allows injecting a mock Redis client creator for testing purposes.
//
// Summary: Injects a mock Redis client creator for unit testing.
//
// Parameters:
//   - creator: func(opts *redis.Options) *redis.Client. The factory function.
func SetRedisClientCreatorForTests(creator func(opts *redis.Options) *redis.Client) {
	redisClientCreator = creator
}

var timeNow = time.Now

// SetTimeNowForTests allows injecting a mock time provider for deterministic testing.
//
// Summary: Overrides the time source for deterministic testing.
//
// Parameters:
//   - nowFunc: func() time.Time. The function returning the current time.
func SetTimeNowForTests(nowFunc func() time.Time) {
	timeNow = nowFunc
}

// RedisLimiter implements a distributed rate limiter backed by Redis.
// It uses a token bucket algorithm to enforce rate limits across multiple service instances,
// ensuring that the configured Requests Per Second (RPS) and burst limits are respected
// regardless of how many server replicas are running.
type RedisLimiter struct {
	client     *redis.Client
	key        string
	rps        float64
	burst      int
	configHash string
}

// NewRedisLimiter creates a new RedisLimiter for the specified service using the provided configuration.
//
// Summary: Initializes a Redis-backed rate limiter for a service.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//   - config: *configv1.RateLimitConfig. The rate limit configuration.
//
// Returns:
//   - *RedisLimiter: The initialized limiter.
//   - error: An error if initialization fails.
func NewRedisLimiter(serviceID string, config *configv1.RateLimitConfig) (*RedisLimiter, error) {
	return NewRedisLimiterWithPartition(serviceID, "", "", config)
}

// NewRedisLimiterWithPartition creates a new RedisLimiter with support for partitioned rate limiting.
//
// Summary: Initializes a partitioned Redis-backed rate limiter (e.g. per-user, per-IP).
//
// Parameters:
//   - serviceID: string. The service identifier.
//   - limitScopeKey: string. The scope identifier (e.g. "user").
//   - partitionKey: string. The partition identifier (e.g. "user:123").
//   - config: *configv1.RateLimitConfig. The configuration.
//
// Returns:
//   - *RedisLimiter: The initialized limiter.
//   - error: An error if Redis config is missing.
func NewRedisLimiterWithPartition(serviceID, limitScopeKey, partitionKey string, config *configv1.RateLimitConfig) (*RedisLimiter, error) {
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

	key := "ratelimit:" + serviceID
	if limitScopeKey != "" {
		key = key + ":" + limitScopeKey
	}
	if partitionKey != "" {
		key = key + ":" + partitionKey
	}

	return &RedisLimiter{
		client: client,
		key:    key,
		rps:    config.GetRequestsPerSecond(),
		burst:  int(config.GetBurst()),
	}, nil
}

// NewRedisLimiterWithClient creates a new RedisLimiter using an existing Redis client.
//
// Summary: Initializes a Redis limiter using a pre-configured Redis client.
//
// Parameters:
//   - client: *redis.Client. The Redis client.
//   - serviceID: string. The service identifier.
//   - limitScopeKey: string. The scope identifier.
//   - partitionKey: string. The partition identifier.
//   - config: *configv1.RateLimitConfig. The configuration.
//
// Returns:
//   - *RedisLimiter: The initialized limiter.
func NewRedisLimiterWithClient(client *redis.Client, serviceID, limitScopeKey, partitionKey string, config *configv1.RateLimitConfig) *RedisLimiter {
	key := "ratelimit:" + serviceID
	if limitScopeKey != "" {
		key = key + ":" + limitScopeKey
	}
	if partitionKey != "" {
		key = key + ":" + partitionKey
	}
	// Calculate config hash
	redisConfig := config.GetRedis()
	configHash := ""
	if redisConfig != nil {
		configHash = redisConfig.GetAddress() + "|" + redisConfig.GetPassword() + "|" + strconv.Itoa(int(redisConfig.GetDb()))
	}
	return &RedisLimiter{
		client:     client,
		key:        key,
		rps:        config.GetRequestsPerSecond(),
		burst:      int(config.GetBurst()),
		configHash: configHash,
	}
}

// RedisRateLimitScript is the Lua script executed atomically in Redis to perform token bucket updates.
// It handles token refill based on time elapsed, checks against burst capacity, and manages
// the expiration of unused keys to prevent memory leaks in Redis.
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

        -- Optimization: Only write EXPIRE if necessary to reduce replication traffic.
        -- TTL command returns -1 if no expiry, -2 if missing (should not happen here).
        local current_ttl = redis.call("TTL", key)
        if current_ttl < (ttl / 2) then
             redis.call("EXPIRE", key, ttl)
        end
        return 1
    end

    return 0
    `

var redisRateLimitScript = redis.NewScript(RedisRateLimitScript)

// Allow checks if a single request is allowed under the current rate limit policy.
//
// Summary: Checks if one token is available in the Redis bucket.
//
// Parameters:
//   - ctx: context.Context. The request context.
//
// Returns:
//   - bool: True if allowed.
//   - error: An error if the Redis operation fails.
func (l *RedisLimiter) Allow(ctx context.Context) (bool, error) {
	return l.AllowN(ctx, 1)
}

// AllowN checks if a request with a specific cost is allowed.
//
// Summary: Checks if 'n' tokens are available in the Redis bucket.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - n: int. The number of tokens to consume.
//
// Returns:
//   - bool: True if allowed.
//   - error: An error if the Redis operation fails.
func (l *RedisLimiter) AllowN(ctx context.Context, n int) (bool, error) {
	now := timeNow().UnixMicro()

	// Use float64 for rate to handle fractional rates
	// Use Run (EVALSHA) for better performance
	cmd := redisRateLimitScript.Run(ctx, l.client, []string{l.key}, l.rps, l.burst, now, n)
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

// Update dynamically updates the rate limit configuration for the running limiter.
//
// Summary: Updates the RPS and burst settings for the limiter.
//
// Parameters:
//   - rps: float64. The new requests per second.
//   - burst: int. The new burst capacity.
func (l *RedisLimiter) Update(rps float64, burst int) {
	l.rps = rps
	l.burst = burst
}

// GetConfigHash returns a hash string representing the underlying Redis configuration.
//
// Summary: Returns a hash of the Redis configuration to detect changes.
//
// Returns:
//   - string: The configuration hash.
func (l *RedisLimiter) GetConfigHash() string {
	return l.configHash
}

// Close terminates the Redis client connection and releases resources.
//
// Summary: Closes the underlying Redis client connection.
//
// Returns:
//   - error: An error if closing fails.
func (l *RedisLimiter) Close() error {
	return l.client.Close()
}
