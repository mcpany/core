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
type RedisStrategy struct {
	// redisClients caches Redis clients per service. Key is serviceID.
	redisClients sync.Map
}

// NewRedisStrategy creates a new RedisStrategy.
//
// Summary: Initializes a new RedisStrategy.
//
// Returns:
//   - *RedisStrategy: A new instance of RedisStrategy.
func NewRedisStrategy() *RedisStrategy {
	return &RedisStrategy{}
}

type cachedRedisClient struct {
	client     *redis.Client
	configHash string
}

// Create creates a new RedisLimiter based on the configuration.
//
// Summary: Creates a Redis-backed rate limiter.
//
// Parameters:
//   - _ : context.Context. Unused.
//   - serviceID: string. The unique identifier of the service.
//   - limitScopeKey: string. The scope of the limit.
//   - partitionKey: string. The partition key.
//   - config: *configv1.RateLimitConfig. The rate limit configuration.
//
// Returns:
//   - Limiter: The created Redis limiter.
//   - error: An error if the Redis configuration is missing or connection fails.
func (s *RedisStrategy) Create(_ context.Context, serviceID, limitScopeKey, partitionKey string, config *configv1.RateLimitConfig) (Limiter, error) {
	if config.GetRedis() == nil {
		return nil, fmt.Errorf("redis config is missing")
	}
	client, err := s.getRedisClient(serviceID, config.GetRedis())
	if err != nil {
		return nil, err
	}
	return NewRedisLimiterWithClient(client, serviceID, limitScopeKey, partitionKey, config), nil
}

func (s *RedisStrategy) getRedisClient(serviceID string, config *bus.RedisBus) (*redis.Client, error) { //nolint:unparam
	configHash := config.GetAddress() + "|" + config.GetPassword() + "|" + strconv.Itoa(int(config.GetDb()))

	if val, ok := s.redisClients.Load(serviceID); ok {
		if cached, ok := val.(*cachedRedisClient); ok {
			if cached.configHash == configHash {
				return cached.client, nil
			}
		}
	}

	opts := &redis.Options{
		Addr:     config.GetAddress(),
		Password: config.GetPassword(),
		DB:       int(config.GetDb()),
	}
	client := redisClientCreator(opts)
	s.redisClients.Store(serviceID, &cachedRedisClient{
		client:     client,
		configHash: configHash,
	})
	return client, nil
}
