// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/mcpany/core/pkg/middleware"
	busproto "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRedisLimiter_Errors(t *testing.T) {
	// Fix time
	fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	middleware.SetTimeNowForTests(func() time.Time { return fixedTime })
	defer middleware.SetTimeNowForTests(time.Now)

	t.Run("Missing Redis Config", func(t *testing.T) {
		rlConfig := configv1.RateLimitConfig_builder{
			IsEnabled: proto.Bool(true),
		}.Build()

		_, err := middleware.NewRedisLimiterWithPartition("service", "", rlConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis config is missing")
	})

	t.Run("Redis Error in Allow", func(t *testing.T) {
		db, redisMock := redismock.NewClientMock()

		rlConfig := configv1.RateLimitConfig_builder{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(10),
			Burst:             proto.Int64(10),
			Redis: busproto.RedisBus_builder{
				Address: proto.String("addr"),
			}.Build(),
		}.Build()

		limiter := middleware.NewRedisLimiterWithClient(db, "service", "", rlConfig)

		redisMock.ExpectEval(
			middleware.RedisRateLimitScript,
			[]string{"ratelimit:service"},
			10.0,                  // rps
			10,                    // burst
			fixedTime.UnixMicro(), // now
			1,                     // cost
		).SetErr(assert.AnError)

		allowed, err := limiter.Allow(context.Background())
		assert.Error(t, err)
		assert.False(t, allowed)
	})

	t.Run("Redis Unexpected Type", func(t *testing.T) {
		db, redisMock := redismock.NewClientMock()

		rlConfig := configv1.RateLimitConfig_builder{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(10),
			Burst:             proto.Int64(10),
		}.Build()

		limiter := middleware.NewRedisLimiterWithClient(db, "service", "", rlConfig)

		redisMock.ExpectEval(
			middleware.RedisRateLimitScript,
			[]string{"ratelimit:service"},
			10.0,
			10,
			fixedTime.UnixMicro(),
			1,
		).SetVal("some string")

		allowed, err := limiter.Allow(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected return type")
		assert.False(t, allowed)
	})

	t.Run("With Partition Key", func(t *testing.T) {
		db, redisMock := redismock.NewClientMock()
		rlConfig := configv1.RateLimitConfig_builder{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(10),
			Burst:             proto.Int64(10),
		}.Build()

		limiter := middleware.NewRedisLimiterWithClient(db, "service", "partition", rlConfig)

		// Expect key to be ratelimit:service:partition
		redisMock.ExpectEval(
			middleware.RedisRateLimitScript,
			[]string{"ratelimit:service:partition"},
			10.0,
			10,
			fixedTime.UnixMicro(),
			1,
		).SetVal(int64(1))

		allowed, err := limiter.Allow(context.Background())
		assert.NoError(t, err)
		assert.True(t, allowed)
	})
}
