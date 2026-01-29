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
// creator: A function that takes Redis options and returns a client instance.
func SetRedisClientCreatorForTests(creator func(opts *redis.Options) *redis.Client) {
	redisClientCreator = creator
}

var timeNow = time.Now

// SetTimeNowForTests allows injecting a mock time provider for deterministic testing.
//
// nowFunc: A function that returns the current time.
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
// It initializes a connection to Redis and sets up the rate limiting parameters.
//
// Parameters:
//   - serviceID: The unique identifier of the service to be rate-limited.
//   - config: The rate limit configuration containing Redis connection details, RPS, and burst settings.
//
// Returns:
//   - A pointer to the initialized RedisLimiter, or nil if an error occurs.
//   - An error if the configuration is invalid or the Redis connection fails.
func NewRedisLimiter(serviceID string, config *configv1.RateLimitConfig) (*RedisLimiter, error) {
	return NewRedisLimiterWithPartition(serviceID, "", "", config)
}

// NewRedisLimiterWithPartition creates a new RedisLimiter with support for partitioned rate limiting.
// This is useful for more granular control, such as per-user or per-IP limits within a service.
//
// Parameters:
//   - serviceID: The unique identifier of the service.
//   - limitScopeKey: An optional key to scope the limit (e.g., "user_id").
//   - partitionKey: An optional key to further partition the limit (e.g., "12345").
//   - config: The rate limit configuration.
//
// Returns:
//   - A pointer to the initialized RedisLimiter, or nil if initialization fails.
//   - An error if the Redis configuration is missing.
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
// This avoids creating a new connection pool if one is already available.
//
// Parameters:
//   - client: The existing Redis client instance.
//   - serviceID: The unique identifier of the service.
//   - limitScopeKey: An optional key to scope the limit.
//   - partitionKey: An optional key to further partition the limit.
//   - config: The rate limit configuration.
//
// Returns:
//   - A pointer to the initialized RedisLimiter.
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
// It decrements the token bucket by 1.
//
// Parameters:
//   - ctx: The context for the request, used for timeouts and cancellation.
//
// Returns:
//   - true if the request is allowed (tokens were available).
//   - false if the request is denied (rate limit exceeded).
//   - An error if the Redis operation fails.
func (l *RedisLimiter) Allow(ctx context.Context) (bool, error) {
	return l.AllowN(ctx, 1)
}

// AllowN checks if a request with a specific cost is allowed.
// It attempts to consume 'n' tokens from the bucket.
//
// Parameters:
//   - ctx: The context for the request.
//   - n: The cost of the request (number of tokens to consume).
//
// Returns:
//   - true if the request is allowed.
//   - false if the request is denied.
//   - An error if the Redis operation fails.
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
// Parameters:
//   - rps: The new requests per second limit.
//   - burst: The new burst capacity.
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
