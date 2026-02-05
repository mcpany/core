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

// ShouldUpdate checks if the limiter needs to be recreated based on the new configuration.
//
// limiter is the existing limiter instance.
// config is the new configuration.
//
// Returns true if the limiter should be recreated.
func (s *RedisStrategy) ShouldUpdate(limiter Limiter, config *configv1.RateLimitConfig) bool {
	rl, ok := limiter.(interface{ GetConfigHash() string })
	if !ok {
		// If existing limiter doesn't support config hash (e.g. was local), we should recreate.
		return true
	}

	redisConfig := config.GetRedis()
	if redisConfig == nil {
		return true
	}

	newConfigHash := redisConfig.GetAddress() + "|" + redisConfig.GetPassword() + "|" + strconv.Itoa(int(redisConfig.GetDb()))
	return rl.GetConfigHash() != newConfigHash
}
