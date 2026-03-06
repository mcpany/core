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

// SetRedisClientCreatorForTests - Auto-generated documentation.
//
// Summary: SetRedisClientCreatorForTests allows injecting a mock Redis client creator for testing purposes.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func SetRedisClientCreatorForTests(creator func(opts *redis.Options) *redis.Client) {
	redisClientCreator = creator
}

// RedisLimiter - Auto-generated documentation.
//
// Summary: RedisLimiter implements a distributed rate limiter backed by Redis.
//
// Fields:
//   - Various fields for RedisLimiter.
type RedisLimiter struct {
	client     *redis.Client
	key        string
	rps        float64
	burst      int
	configHash string
}

// NewRedisLimiter creates a new RedisLimiter for the specified service using the provided configuration.
// It initializes a connection to Redis and sets up the rate limiting parameters.
//
// Summary: Initializes a new Redis-backed rate limiter.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service to be rate-limited.
//   - config: *configv1.RateLimitConfig. The configuration containing Redis connection details, RPS, and burst settings.
//
// Returns:
//   - *RedisLimiter: The initialized RedisLimiter, or nil if an error occurs.
//   - error: An error if the configuration is invalid or the Redis connection fails.
//
// Side Effects:
//   - Creates a new Redis connection.
func NewRedisLimiter(serviceID string, config *configv1.RateLimitConfig) (*RedisLimiter, error) {
	return NewRedisLimiterWithPartition(serviceID, "", "", config)
}

// NewRedisLimiterWithPartition creates a new RedisLimiter with support for partitioned rate limiting.
// This is useful for more granular control, such as per-user or per-IP limits within a service.
//
// Summary: Initializes a Redis-backed rate limiter with partitioning support.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//   - limitScopeKey: string. An optional key to scope the limit (e.g., "user_id").
//   - partitionKey: string. An optional key to further partition the limit (e.g., "12345").
//   - config: *configv1.RateLimitConfig. The rate limit configuration.
//
// Returns:
//   - *RedisLimiter: The initialized limiter.
//   - error: An error if Redis config is missing.
//
// Errors:
//   - Returns "redis config is missing" if config is incomplete.
//
// Side Effects:
//   - Creates a new Redis connection.
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

// NewRedisLimiterWithClient creates a new RedisLimiter using an existing Redis client. This avoids creating a new connection pool if one is already available. Summary: Initializes a RedisLimiter reusing an existing Redis client. Parameters: - client: *redis.Client. The existing Redis client instance. - serviceID: string. The unique identifier of the service. - limitScopeKey: string. An optional key to scope the limit. - partitionKey: string. An optional key to further partition the limit. - config: *configv1.RateLimitConfig. The rate limit configuration. Returns: - *RedisLimiter: The initialized limiter.
//
// Summary: NewRedisLimiterWithClient creates a new RedisLimiter using an existing Redis client. This avoids creating a new connection pool if one is already available. Summary: Initializes a RedisLimiter reusing an existing Redis client. Parameters: - client: *redis.Client. The existing Redis client instance. - serviceID: string. The unique identifier of the service. - limitScopeKey: string. An optional key to scope the limit. - partitionKey: string. An optional key to further partition the limit. - config: *configv1.RateLimitConfig. The rate limit configuration. Returns: - *RedisLimiter: The initialized limiter.
//
// Parameters:
//   - client (*redis.Client): The client parameter used in the operation.
//   - _ (serviceID): An unnamed parameter of type serviceID.
//   - _ (limitScopeKey): An unnamed parameter of type limitScopeKey.
//   - partitionKey (string): The partition key parameter used in the operation.
//   - config (*configv1.RateLimitConfig): The configuration settings to be applied.
//
// Returns:
//   - (*RedisLimiter): The resulting RedisLimiter object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
// It decrements the token bucket by 1.
//
// Summary: Checks if a single request is allowed.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//
// Returns:
//   - bool: true if the request is allowed.
//   - error: An error if the Redis operation fails.
//
// Side Effects:
//   - Executes a Lua script on Redis to atomically consume tokens.
func (l *RedisLimiter) Allow(ctx context.Context) (bool, error) {
	return l.AllowN(ctx, 1)
}

// AllowN checks if a request with a specific cost is allowed.
// It attempts to consume 'n' tokens from the bucket.
//
// Summary: Checks if a request with cost N is allowed.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - n: int. The cost of the request.
//
// Returns:
//   - bool: true if the request is allowed.
//   - error: An error if the Redis operation fails.
//
// Side Effects:
//   - Executes a Lua script on Redis to atomically consume tokens.
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

// Update - Auto-generated documentation.
//
// Summary: Update dynamically updates the rate limit configuration for the running limiter.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (l *RedisLimiter) Update(rps float64, burst int) {
	l.rps = rps
	l.burst = burst
}

// GetConfigHash - Auto-generated documentation.
//
// Summary: GetConfigHash returns a hash string representing the underlying Redis configuration.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (l *RedisLimiter) GetConfigHash() string {
	return l.configHash
}

// Close - Auto-generated documentation.
//
// Summary: Close terminates the Redis client connection and releases resources.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (l *RedisLimiter) Close() error {
	return l.client.Close()
}
