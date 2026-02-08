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
// Summary: allows injecting a mock Redis client creator for testing purposes.
//
// Parameters:
//   - creator: func(opts *redis.Options) *redis.Client. The creator.
//
// Returns:
//   None.
func SetRedisClientCreatorForTests(creator func(opts *redis.Options) *redis.Client) {
	redisClientCreator = creator
}

var timeNow = time.Now

// SetTimeNowForTests allows injecting a mock time provider for deterministic testing.
//
// Summary: allows injecting a mock time provider for deterministic testing.
//
// Parameters:
//   - nowFunc: func() time.Time. The nowFunc.
//
// Returns:
//   None.
func SetTimeNowForTests(nowFunc func() time.Time) {
	timeNow = nowFunc
}

// RedisLimiter implements a distributed rate limiter backed by Redis.
//
// Summary: implements a distributed rate limiter backed by Redis.
type RedisLimiter struct {
	client     *redis.Client
	key        string
	rps        float64
	burst      int
	configHash string
}

// NewRedisLimiter creates a new RedisLimiter for the specified service using the provided configuration.
//
// Summary: creates a new RedisLimiter for the specified service using the provided configuration.
//
// Parameters:
//   - serviceID: string. The serviceID.
//   - config: *configv1.RateLimitConfig. The config.
//
// Returns:
//   - *RedisLimiter: The *RedisLimiter.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func NewRedisLimiter(serviceID string, config *configv1.RateLimitConfig) (*RedisLimiter, error) {
	return NewRedisLimiterWithPartition(serviceID, "", "", config)
}

// NewRedisLimiterWithPartition creates a new RedisLimiter with support for partitioned rate limiting.
//
// Summary: creates a new RedisLimiter with support for partitioned rate limiting.
//
// Parameters:
//   - serviceID: string. The serviceID.
//   - limitScopeKey: string. The limitScopeKey.
//   - partitionKey: string. The partitionKey.
//   - config: *configv1.RateLimitConfig. The config.
//
// Returns:
//   - *RedisLimiter: The *RedisLimiter.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: creates a new RedisLimiter using an existing Redis client.
//
// Parameters:
//   - client: *redis.Client. The client.
//   - serviceID: string. The serviceID.
//   - limitScopeKey: string. The limitScopeKey.
//   - partitionKey: string. The partitionKey.
//   - config: *configv1.RateLimitConfig. The config.
//
// Returns:
//   - *RedisLimiter: The *RedisLimiter.
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
// Summary: checks if a single request is allowed under the current rate limit policy.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - bool: The bool.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (l *RedisLimiter) Allow(ctx context.Context) (bool, error) {
	return l.AllowN(ctx, 1)
}

// AllowN checks if a request with a specific cost is allowed.
//
// Summary: checks if a request with a specific cost is allowed.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - n: int. The n.
//
// Returns:
//   - bool: The bool.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: dynamically updates the rate limit configuration for the running limiter.
//
// Parameters:
//   - rps: float64. The rps.
//   - burst: int. The burst.
//
// Returns:
//   None.
func (l *RedisLimiter) Update(rps float64, burst int) {
	l.rps = rps
	l.burst = burst
}

// GetConfigHash returns a hash string representing the underlying Redis configuration.
//
// Summary: returns a hash string representing the underlying Redis configuration.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The string.
func (l *RedisLimiter) GetConfigHash() string {
	return l.configHash
}

// Close terminates the Redis client connection and releases resources.
//
// Summary: terminates the Redis client connection and releases resources.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (l *RedisLimiter) Close() error {
	return l.client.Close()
}
