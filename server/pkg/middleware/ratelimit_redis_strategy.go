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
// Returns the result.
func NewRedisStrategy() *RedisStrategy {
	return &RedisStrategy{}
}

type cachedRedisClient struct {
	client     *redis.Client
	configHash string
}

// Create creates a new RedisLimiter.
//
// _ is an unused parameter.
// serviceID is the serviceID.
// limitScopeKey is the limitScopeKey.
// partitionKey is the partitionKey.
// config holds the configuration settings.
//
// Returns the result.
// Returns an error if the operation fails.
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

func (s *RedisStrategy) getRedisClient(_ string, config *bus.RedisBus) (*redis.Client, error) { //nolint:unparam
	configHash := config.GetAddress() + "|" + config.GetPassword() + "|" + strconv.Itoa(int(config.GetDb()))

	// ⚡ BOLT: Key by configHash instead of serviceID to share connection pools across services.
	// Randomized Selection from Top 5 High-Impact Targets
	if val, ok := s.redisClients.Load(configHash); ok {
		if cached, ok := val.(*cachedRedisClient); ok {
			return cached.client, nil
		}
	}

	opts := &redis.Options{
		Addr:     config.GetAddress(),
		Password: config.GetPassword(),
		DB:       int(config.GetDb()),
	}
	client := redisClientCreator(opts)
	s.redisClients.Store(configHash, &cachedRedisClient{
		client:     client,
		configHash: configHash,
	})
	return client, nil
}
