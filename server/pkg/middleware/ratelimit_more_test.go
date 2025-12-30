// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/mcpany/core/pkg/tool"
	busproto "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockLimiter for checkLimit test
type MockLimiter struct {
	mock.Mock
}

func (m *MockLimiter) Allow(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockLimiter) AllowN(ctx context.Context, n int) (bool, error) {
	args := m.Called(ctx, n)
	return args.Bool(0), args.Error(1)
}

func (m *MockLimiter) Update(rps float64, burst int) {
	m.Called(rps, burst)
}

func TestCheckLimitFailOpen(t *testing.T) {
	m := &RateLimitMiddleware{}
	ctx := context.Background()
	req := &tool.ExecutionRequest{}
	config := &configv1.RateLimitConfig{}

	mockLimiter := &MockLimiter{}
	mockLimiter.On("AllowN", ctx, 1).Return(false, errors.New("redis error"))

	// Should return nil (allow) on error
	err := m.checkLimit(ctx, mockLimiter, config, req)
	assert.NoError(t, err)
}

func TestGetLimiterRedisConfigMissing(t *testing.T) {
	m := NewRateLimitMiddleware(nil)
	ctx := context.Background()
	serviceID := "service1"
	scope := "scope"

	storageRedis := configv1.RateLimitConfig_STORAGE_REDIS
	config := &configv1.RateLimitConfig{
		Storage: &storageRedis,
		// Redis config missing
	}

	_, err := m.getLimiter(ctx, serviceID, scope, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis config is missing")
}

func TestNewRedisLimiterWithPartition(t *testing.T) {
	config := &configv1.RateLimitConfig{}
	_, err := NewRedisLimiterWithPartition("s1", "p1", config)
	assert.Error(t, err) // missing redis config
}

func TestRedisLimiterAllowError(t *testing.T) {
	db, mockRedis := redismock.NewClientMock()

	middlewareSetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return db
	})
	defer middlewareSetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return redis.NewClient(opts)
	})

	redisBus := busproto.RedisBus_builder{
		Address: proto.String("localhost:6379"),
	}.Build()

	config := &configv1.RateLimitConfig{
		IsEnabled:         proto.Bool(true),
		RequestsPerSecond: proto.Float64(10),
		Burst:             proto.Int64(10),
		Redis:             redisBus,
	}

	rl, err := NewRedisLimiterWithPartition("s1", "p1", config)
	assert.NoError(t, err)

	ctx := context.Background()

	// Mock Script Run Error
	mockRedis.ExpectEvalSha(mock.Anything, []string{"ratelimit:s1:p1"}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		SetErr(errors.New("connection failed"))

	allowed, err := rl.Allow(ctx)
	assert.False(t, allowed)
	assert.Error(t, err)
}

func TestRedisLimiterAllowUnexpectedType(t *testing.T) {
	db, mockRedis := redismock.NewClientMock()

	middlewareSetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return db
	})
	defer middlewareSetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return redis.NewClient(opts)
	})

	// Set fixed time
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	middlewareSetTimeNowForTests(func() time.Time {
		return fixedTime
	})
	defer middlewareSetTimeNowForTests(time.Now)

	redisBus := busproto.RedisBus_builder{
		Address: proto.String("localhost:6379"),
	}.Build()

	config := &configv1.RateLimitConfig{
		IsEnabled:         proto.Bool(true),
		RequestsPerSecond: proto.Float64(10),
		Burst:             proto.Int64(10),
		Redis:             redisBus,
	}

	rl, err := NewRedisLimiterWithPartition("s1", "p1", config)
	assert.NoError(t, err)

	ctx := context.Background()

	// Use explicit hash which matches the script
	s := redis.NewScript(RedisRateLimitScript)

	// Mock Script Run Success but unexpected return type (e.g., string instead of int64)
	mockRedis.ExpectEvalSha(s.Hash(), []string{"ratelimit:s1:p1"}, 10.0, 10, fixedTime.UnixMicro(), 1).
		SetVal("unexpected")

	allowed, err := rl.Allow(ctx)
	assert.False(t, allowed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected return type")
}

func middlewareSetRedisClientCreatorForTests(creator func(opts *redis.Options) *redis.Client) {
	SetRedisClientCreatorForTests(creator)
}

func middlewareSetTimeNowForTests(nowFunc func() time.Time) {
	SetTimeNowForTests(nowFunc)
}
