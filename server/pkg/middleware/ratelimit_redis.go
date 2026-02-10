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
// Summary: Injects a mock Redis client creator.
//
// Parameters:
//   - creator: func(opts *redis.Options) *redis.Client. The mock creator function.
func SetRedisClientCreatorForTests(creator func(opts *redis.Options) *redis.Client) {
	redisClientCreator = creator
}

// RedisLimiter implements a distributed rate limiter backed by Redis.
//
// Summary: Distributed rate limiter using Redis token bucket.
//
// Fields:
//   - client: *redis.Client. The Redis client.
//   - key: string. The rate limit key in Redis.
//   - rps: float64. Requests per second.
//   - burst: int. Burst capacity.
//   - configHash: string. Hash of the Redis configuration for change detection.
type RedisLimiter struct {
	client     *redis.Client
	key        string
	rps        float64
	burst      int
	configHash string
}

// NewRedisLimiter creates a new RedisLimiter for the specified service using the provided configuration.
//
// Summary: Creates a new Redis-backed rate limiter.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service to be rate-limited.
//   - config: *configv1.RateLimitConfig. The rate limit configuration containing Redis connection details, RPS, and burst settings.
//
// Returns:
//   - *RedisLimiter: A pointer to the initialized RedisLimiter, or nil if an error occurs.
//   - error: An error if the configuration is invalid or the Redis connection fails.
func NewRedisLimiter(serviceID string, config *configv1.RateLimitConfig) (*RedisLimiter, error) {
	return NewRedisLimiterWithPartition(serviceID, "", "", config)
}

// NewRedisLimiterWithPartition creates a new RedisLimiter with support for partitioned rate limiting.
//
// Summary: Creates a partitioned Redis rate limiter.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//   - limitScopeKey: string. An optional key to scope the limit (e.g., "user_id").
//   - partitionKey: string. An optional key to further partition the limit (e.g., "12345").
//   - config: *configv1.RateLimitConfig. The rate limit configuration.
//
// Returns:
//   - *RedisLimiter: A pointer to the initialized RedisLimiter, or nil if initialization fails.
//   - error: An error if the Redis configuration is missing.
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
// Summary: Creates a Redis limiter reusing an existing client.
//
// Parameters:
//   - client: *redis.Client. The existing Redis client instance.
//   - serviceID: string. The unique identifier of the service.
//   - limitScopeKey: string. An optional key to scope the limit.
//   - partitionKey: string. An optional key to further partition the limit.
//   - config: *configv1.RateLimitConfig. The rate limit configuration.
//
// Returns:
//   - *RedisLimiter: A pointer to the initialized RedisLimiter.
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
//
// Summary: Lua script for atomic token bucket in Redis.
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
// Summary: Checks if one token is available.
//
// Parameters:
//   - ctx: context.Context. The context for the request, used for timeouts and cancellation.
//
// Returns:
//   - bool: true if the request is allowed (tokens were available).
//   - error: An error if the Redis operation fails.
func (l *RedisLimiter) Allow(ctx context.Context) (bool, error) {
	return l.AllowN(ctx, 1)
}

// AllowN checks if a request with a specific cost is allowed.
//
// Summary: Checks if N tokens are available.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - n: int. The cost of the request (number of tokens to consume).
//
// Returns:
//   - bool: true if the request is allowed.
//   - error: An error if the Redis operation fails.
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
// Summary: Updates rate and burst.
//
// Parameters:
//   - rps: float64. The new requests per second limit.
//   - burst: int. The new burst capacity.
func (l *RedisLimiter) Update(rps float64, burst int) {
	l.rps = rps
	l.burst = burst
}

// GetConfigHash returns a hash string representing the underlying Redis configuration.
//
// Summary: Returns config hash for change detection.
//
// Returns:
//   - string: The configuration hash string.
func (l *RedisLimiter) GetConfigHash() string {
	return l.configHash
}

// Close terminates the Redis client connection and releases resources.
//
// Summary: Closes the Redis client.
//
// Returns:
//   - error: An error if closing the client fails.
func (l *RedisLimiter) Close() error {
	return l.client.Close()
}
