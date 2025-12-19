// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	busproto "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRateLimitMiddleware(t *testing.T) {
	const successResult = "success"
	t.Run("rate limit allowed", func(t *testing.T) {
		rlConfig := &configv1.RateLimitConfig{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(10),
			Burst:             proto.Int64(1),
		}

		rlMiddleware, err := middleware.NewRateLimitMiddleware("service", rlConfig)
		assert.NoError(t, err)

		req := &tool.ExecutionRequest{
			ToolName: "service.test-tool",
		}

		nextCalled := false
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return successResult, nil
		}

		result, err := rlMiddleware.Execute(context.Background(), req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)
		assert.True(t, nextCalled)
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		rlConfig := &configv1.RateLimitConfig{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(1),
			Burst:             proto.Int64(1),
		}

		rlMiddleware, err := middleware.NewRateLimitMiddleware("service", rlConfig)
		assert.NoError(t, err)

		req := &tool.ExecutionRequest{ToolName: "tool"}
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			return successResult, nil
		}

		// First request allowed
		_, err = rlMiddleware.Execute(context.Background(), req, next)
		assert.NoError(t, err)

		// Second request blocked (1 RPS, Burst 1 consumed)
		_, err = rlMiddleware.Execute(context.Background(), req, next)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")
	})

	t.Run("rate limit disabled", func(t *testing.T) {
		rlConfig := &configv1.RateLimitConfig{
			IsEnabled: proto.Bool(false),
		}

		mw, err := middleware.NewRateLimitMiddleware("service", rlConfig)
		assert.NoError(t, err)
		assert.Nil(t, mw)
	})

	t.Run("redis rate limit", func(t *testing.T) {
		db, mockRedis := redismock.NewClientMock()
		middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
			return db
		})
		defer middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
			return redis.NewClient(opts)
		})

		fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		middleware.SetTimeNowForTests(func() time.Time {
			return fixedTime
		})
		defer middleware.SetTimeNowForTests(time.Now)

		storageRedis := configv1.RateLimitConfig_STORAGE_REDIS
		rlConfig := &configv1.RateLimitConfig{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(10),
			Burst:             proto.Int64(10),
			Storage:           &storageRedis,
			Redis: &busproto.RedisBus{
				Address: proto.String("localhost:6379"),
			},
		}

		mw, err := middleware.NewRateLimitMiddleware("service", rlConfig)
		assert.NoError(t, err)

		mockRedis.ExpectEval(
			middleware.RedisRateLimitScript,
			[]string{"ratelimit:service"},
			10.0,
			10,
			fixedTime.UnixMicro(),
			1,
		).SetVal(int64(1))

		req := &tool.ExecutionRequest{ToolName: "tool"}
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			return successResult, nil
		}

		res, err := mw.Execute(context.Background(), req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, res)
		assert.NoError(t, mockRedis.ExpectationsWereMet())
	})
}
