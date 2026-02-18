// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"strconv"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/redis/go-redis/v9"
)

var redisClientCreator = redis.NewClient

// SetRedisClientCreatorForTests allows injecting a mock Redis client creator for testing purposes.
//
// Summary: Overrides the Redis client creation function for tests.
//
// Parameters:
//   - creator: func(opts *redis.Options) *redis.Client. The mock creator function.
func SetRedisClientCreatorForTests(creator func(opts *redis.Options) *redis.Client) {
	redisClientCreator = creator
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
// Summary: Creates a Redis-backed rate limiter for a service.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//   - config: *configv1.RateLimitConfig. The rate limit configuration.
//
// Returns:
//   - *RedisLimiter: The initialized limiter.
//   - error: Error if initialization fails.
func NewRedisLimiter(serviceID string, config *configv1.RateLimitConfig) (*RedisLimiter, error) {
	return NewRedisLimiterWithPartition(serviceID, "", "", config)
}

// NewRedisLimiterWithPartition creates a new RedisLimiter with support for partitioned rate limiting.
//
// Summary: Creates a partitioned Redis-backed rate limiter.
//
// Parameters:
//   - serviceID: string. The service ID.
//   - limitScopeKey: string. The scope key (e.g. "user_id").
//   - partitionKey: string. The partition key (e.g. "123").
//   - config: *configv1.RateLimitConfig. The configuration.
//
// Returns:
//   - *RedisLimiter: The initialized limiter.
//   - error: Error if initialization fails.
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
// Summary: Creates a Redis-backed rate limiter using an existing client.
//
// Parameters:
//   - client: *redis.Client. The existing Redis client.
//   - serviceID: string. The service ID.
//   - limitScopeKey: string. The scope key.
//   - partitionKey: string. The partition key.
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
    -- ⚡ BOLT: Use server time to prevent clock skew issues in distributed systems.
    -- Randomized Selection from Top 5 High-Impact Targets
    redis.replicate_commands()
    local time = redis.call("TIME")
    local now = tonumber(time[1]) * 1000000 + tonumber(time[2])

    local key = KEYS[1]
    local rate = tonumber(ARGV[1])
    local burst = tonumber(ARGV[2])
    -- ARGV[3] was 'now', now it is 'cost'
    local cost = tonumber(ARGV[3])

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
// Summary: Checks permission for a single request (cost 1).
//
// Parameters:
//   - ctx: context.Context. The request context.
//
// Returns:
//   - bool: True if allowed, false otherwise.
//   - error: Error if Redis operation fails.
func (l *RedisLimiter) Allow(ctx context.Context) (bool, error) {
	return l.AllowN(ctx, 1)
}

// AllowN checks if a request with a specific cost is allowed.
//
// Summary: Checks permission for a request with a specific cost.
//
// Parameters:
//   - ctx: context.Context. The request context.
//   - n: int. The cost of the request.
//
// Returns:
//   - bool: True if allowed, false otherwise.
//   - error: Error if Redis operation fails.
func (l *RedisLimiter) AllowN(ctx context.Context, n int) (bool, error) {
	// ⚡ BOLT: Use server time to prevent clock skew issues in distributed systems.
	// Randomized Selection from Top 5 High-Impact Targets
	// now is fetched inside the script using redis.call("TIME")
	cmd := redisRateLimitScript.Run(ctx, l.client, []string{l.key}, l.rps, l.burst, n)
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
// Summary: Updates the rate limit settings dynamically.
//
// Parameters:
//   - rps: float64. The new requests per second limit.
//   - burst: int. The new burst capacity.
func (l *RedisLimiter) Update(rps float64, burst int) {
	l.rps = rps
	l.burst = burst
}

// GetConfigHash returns a hash string representing the underlying Redis configuration.
// This is used to detect configuration changes that might require a client reconnection.
//
// Returns:
//   - The configuration hash string.
func (l *RedisLimiter) GetConfigHash() string {
	return l.configHash
}

// Close terminates the Redis client connection and releases resources.
//
// Returns:
//   - An error if closing the client fails.
func (l *RedisLimiter) Close() error {
	return l.client.Close()
}
