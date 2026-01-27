// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisStrategy_Create(t *testing.T) {
	db, _ := redismock.NewClientMock()

	originalCreator := redisClientCreator
	defer func() { redisClientCreator = originalCreator }()
	SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return db
	})

	strategy := NewRedisStrategy()

	rBus := &bus.RedisBus{}
	rBus.SetAddress("localhost:6379")
	cfg := &configv1.RateLimitConfig{
		RequestsPerSecond: 10,
		Burst:             10,
		Redis:             rBus,
	}

	limiter, err := strategy.Create(context.Background(), "service1", "", "partition", cfg)
	assert.NoError(t, err)
	assert.NotNil(t, limiter)
	assert.IsType(t, &RedisLimiter{}, limiter)
}

func TestRedisStrategy_Create_MissingConfig(t *testing.T) {
	strategy := NewRedisStrategy()

	// Missing Redis config
	cfg := &configv1.RateLimitConfig{
		RequestsPerSecond: 10,
		Burst:             10,
		Redis:             nil,
	}

	limiter, err := strategy.Create(context.Background(), "service1", "", "partition", cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis config is missing")
	assert.Nil(t, limiter)
}

func TestRedisStrategy_ClientCaching(t *testing.T) {
	db, _ := redismock.NewClientMock()

	originalCreator := redisClientCreator
	defer func() { redisClientCreator = originalCreator }()

	createCount := 0
	SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		createCount++
		return db
	})

	strategy := NewRedisStrategy()

	rBus := &bus.RedisBus{}
	rBus.SetAddress("localhost:6379")
	cfg := &configv1.RateLimitConfig{
		RequestsPerSecond: 10,
		Burst:             10,
		Redis:             rBus,
	}

	// 1. Create first limiter
	_, err := strategy.Create(context.Background(), "service1", "", "partition", cfg)
	assert.NoError(t, err)
	assert.Equal(t, 1, createCount)

	// 2. Create second limiter with same serviceID and config -> reuse client
	_, err = strategy.Create(context.Background(), "service1", "", "partition2", cfg)
	assert.NoError(t, err)
	assert.Equal(t, 1, createCount)

	// 3. Create limiter for different serviceID -> new client
	_, err = strategy.Create(context.Background(), "service2", "", "partition", cfg)
	assert.NoError(t, err)
	assert.Equal(t, 2, createCount)

	// 4. Create limiter for service1 but different config
	rBus2 := &bus.RedisBus{}
	rBus2.SetAddress("localhost:6380")
	cfg2 := &configv1.RateLimitConfig{
		RequestsPerSecond: 10,
		Burst:             10,
		Redis:             rBus2,
	}

	// Since getRedisClient checks configHash of cached client, it should create new one.
	_, err = strategy.Create(context.Background(), "service1", "", "partition", cfg2)
	assert.NoError(t, err)
	assert.Equal(t, 3, createCount)
}
