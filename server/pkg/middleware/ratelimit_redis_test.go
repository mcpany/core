package middleware_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	busproto "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRedisLimiter(t *testing.T) {
	db, mockRedis := redismock.NewClientMock()
	middleware.SetRedisClientCreatorForTests(func(_ *redis.Options) *redis.Client {
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

	t.Run("NewRedisLimiter missing config", func(t *testing.T) {
		config := configv1.RateLimitConfig_builder{}.Build()
		_, err := middleware.NewRedisLimiter("service", config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis config is missing")
	})

	t.Run("Update", func(t *testing.T) {
		config := configv1.RateLimitConfig_builder{
			Redis:             busproto.RedisBus_builder{Address: proto.String("addr")}.Build(),
			RequestsPerSecond: 10,
			Burst:             10,
		}.Build()
		l, err := middleware.NewRedisLimiter("service", config)
		assert.NoError(t, err)

		l.Update(20, 20)

		s := redis.NewScript(middleware.RedisRateLimitScript)
		mockRedis.ExpectEvalSha(s.Hash(), []string{"ratelimit:service"}, 20.0, 20, fixedTime.UnixMicro(), 1).SetVal(int64(1))
		allowed, err := l.Allow(context.Background())
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.NoError(t, mockRedis.ExpectationsWereMet())
	})

	t.Run("Allow error", func(t *testing.T) {
		config := configv1.RateLimitConfig_builder{
			Redis:             busproto.RedisBus_builder{Address: proto.String("addr")}.Build(),
			RequestsPerSecond: 10,
			Burst:             10,
		}.Build()
		l, _ := middleware.NewRedisLimiter("service", config)

		s := redis.NewScript(middleware.RedisRateLimitScript)
		mockRedis.ExpectEvalSha(
			s.Hash(),
			[]string{"ratelimit:service"},
			10.0,
			10,
			fixedTime.UnixMicro(),
			1,
		).SetErr(errors.New("redis error"))
		allowed, err := l.Allow(context.Background())
		assert.Error(t, err)
		assert.False(t, allowed)
	})

	t.Run("Allow unexpected type", func(t *testing.T) {
		config := configv1.RateLimitConfig_builder{
			Redis:             busproto.RedisBus_builder{Address: proto.String("addr")}.Build(),
			RequestsPerSecond: 10,
			Burst:             10,
		}.Build()
		l, _ := middleware.NewRedisLimiter("service", config)

		s := redis.NewScript(middleware.RedisRateLimitScript)
		mockRedis.ExpectEvalSha(
			s.Hash(),
			[]string{"ratelimit:service"},
			10.0,
			10,
			fixedTime.UnixMicro(),
			1,
		).SetVal("not-int")
		allowed, err := l.Allow(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected return type")
		assert.False(t, allowed)
	})

	t.Run("Close", func(t *testing.T) {
		config := configv1.RateLimitConfig_builder{
			Redis: busproto.RedisBus_builder{Address: proto.String("addr")}.Build(),
		}.Build()
		l, _ := middleware.NewRedisLimiter("service", config)

		err := l.Close()
		assert.NoError(t, err)
	})
}
