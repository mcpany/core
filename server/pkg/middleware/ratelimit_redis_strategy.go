// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/redis/go-redis/v9"
)

// RedisStrategy implements RateLimitStrategy for Redis-based rate limiting.
//
// Summary: Strategy for creating Redis-backed distributed rate limiters.
type RedisStrategy struct {
	// redisClients caches Redis clients per config. Key is configHash.
	redisClients sync.Map
}

// NewRedisStrategy creates a new RedisStrategy.
//
// Summary: Initializes a new RedisStrategy.
//
// Returns:
//   - *RedisStrategy: The initialized strategy.
func NewRedisStrategy() *RedisStrategy {
	return &RedisStrategy{}
}

// Create creates a new RedisLimiter.
//
// Summary: Creates a new Redis-backed rate limiter.
//
// Parameters:
//   - _: context.Context. Unused.
//   - serviceID: string. The service identifier.
//   - limitScopeKey: string. The scope key for the limit.
//   - partitionKey: string. The partition key for the limit.
//   - config: *configv1.RateLimitConfig. The rate limit configuration.
//
// Returns:
//   - Limiter: The created RedisLimiter.
//   - error: An error if the Redis configuration is missing.
//
// Errors:
//   - Returns "redis config is missing" if the config does not contain Redis settings.
//
// Side Effects:
//   - Establishes or reuses a Redis connection.
func (s *RedisStrategy) Create(_ context.Context, serviceID, limitScopeKey, partitionKey string, config *configv1.RateLimitConfig) (Limiter, error) {
	if config.GetRedis() == nil {
		return nil, fmt.Errorf("redis config is missing")
	}
	// ⚡ BOLT: Key Redis clients by config hash instead of serviceID.
	// Randomized Selection from Top 5 High-Impact Targets.
	// This ensures that multiple services sharing the same Redis config use the same connection pool.
	client := s.getRedisClient(config.GetRedis())
	return NewRedisLimiterWithClient(client, serviceID, limitScopeKey, partitionKey, config), nil
}

func (s *RedisStrategy) getRedisClient(config *bus.RedisBus) *redis.Client {
	configHash := config.GetAddress() + "|" + config.GetPassword() + "|" + strconv.Itoa(int(config.GetDb()))

	// Fast path: Check if client exists
	if val, ok := s.redisClients.Load(configHash); ok {
		if client, ok := val.(*redis.Client); ok {
			return client
		}
	}

	// Slow path: Create new client and use LoadOrStore to handle race conditions
	opts := &redis.Options{
		Addr:     config.GetAddress(),
		Password: config.GetPassword(),
		DB:       int(config.GetDb()),
	}
	newClient := redisClientCreator(opts)

	actual, loaded := s.redisClients.LoadOrStore(configHash, newClient)
	if loaded {
		// Another goroutine created the client first. Close our redundant one.
		_ = newClient.Close()
		return actual.(*redis.Client)
	}

	return newClient
}
