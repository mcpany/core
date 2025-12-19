// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisLimiter implements a rate limiter using Redis.
type RedisLimiter struct {
	client    *redis.Client
	serviceID string
	rps       float64
	burst     int
}

// NewRedisLimiter creates a new RedisLimiter.
func NewRedisLimiter(client *redis.Client, serviceID string, rps float64, burst int) *RedisLimiter {
	return &RedisLimiter{
		client:    client,
		serviceID: serviceID,
		rps:       rps,
		burst:     burst,
	}
}

// Allow checks if the request is allowed.
func (l *RedisLimiter) Allow(ctx context.Context) (bool, error) {
	// Lua script for token bucket algorithm
	// KEYS[1]: tokens_key
	// KEYS[2]: timestamp_key
	// ARGV[1]: rate (tokens/sec)
	// ARGV[2]: capacity (burst)
	script := redis.NewScript(`
		local tokens_key = KEYS[1]
		local timestamp_key = KEYS[2]
		local rate = tonumber(ARGV[1])
		local capacity = tonumber(ARGV[2])

		local time = redis.call('TIME')
		local now = tonumber(time[1]) + (tonumber(time[2]) / 1000000)

		local last_tokens = tonumber(redis.call("get", tokens_key))
		if last_tokens == nil then
			last_tokens = capacity
		end

		local last_refreshed = tonumber(redis.call("get", timestamp_key))
		if last_refreshed == nil then
			last_refreshed = 0
		end

		local delta = math.max(0, now - last_refreshed)
		local filled_tokens = math.min(capacity, last_tokens + (delta * rate))
		local allowed = filled_tokens >= 1
		local new_tokens = filled_tokens

		if allowed then
			new_tokens = filled_tokens - 1
		end

		-- Calculate TTL to avoid keeping keys forever
		-- If rate is very low, fill_time can be high.
		local ttl = 60
		if rate > 0 then
			local fill_time = capacity / rate
			ttl = math.ceil(fill_time * 2)
			if ttl < 60 then ttl = 60 end
		end

		redis.call("setex", tokens_key, ttl, new_tokens)
		redis.call("setex", timestamp_key, ttl, now)

		if allowed then
			return 1
		else
			return 0
		end
	`)

	tokensKey := fmt.Sprintf("ratelimit:{%s}:tokens", l.serviceID)
	timestampKey := fmt.Sprintf("ratelimit:{%s}:ts", l.serviceID)

	res, err := script.Run(ctx, l.client, []string{tokensKey, timestampKey}, l.rps, l.burst).Result()
	if err != nil {
		return false, err
	}

	allowedInt, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected result type from redis script: %T", res)
	}

	return allowedInt == 1, nil
}

// UpdateConfig updates the limiter configuration.
func (l *RedisLimiter) UpdateConfig(rps float64, burst int) {
	l.rps = rps
	l.burst = burst
}

// Close closes the Redis client.
func (l *RedisLimiter) Close() error {
	return l.client.Close()
}
