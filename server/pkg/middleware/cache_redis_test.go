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
	routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestCachingMiddleware_Redis(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)

	db, mockRedis := redismock.NewClientMock()
	middleware.SetCacheRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return db
	})
	// Reset to default (redis.NewClient) instead of nil to avoid panics if other tests run in parallel
	// But defer is LIFO, so it runs at end of test.
	// Since we don't have access to original, we can just set it to redis.NewClient?
	// But redis.NewClient is a function.
	// Actually for now just nil might be unsafe if code doesn't check for nil.
	// But my code: var cacheRedisClientCreator = redis.NewClient
	// So I should set it back to redis.NewClient.
	defer middleware.SetCacheRedisClientCreatorForTests(redis.NewClient)

	m := middleware.NewCachingMiddleware(mockToolManager)

	req := &tool.ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: []byte("{}"),
	}

	// Cache config
	// Using builder for RedisBus because it might be opaque
	redisBus := busproto.RedisBus_builder{
		Address: proto.String("localhost:6379"),
	}.Build()

	cacheConfig := configv1.CacheConfig_builder{
		IsEnabled: proto.Bool(true),
		Storage:   configv1.CacheConfig_STORAGE_REDIS.Enum(),
		Redis:     redisBus,
		Ttl:       durationpb.New(1 * time.Minute),
	}.Build()

	mockTool := &tool.MockTool{
		ToolFunc: func() *routerv1.Tool {
			return &routerv1.Tool{
				Name:      proto.String("test-tool"),
				ServiceId: proto.String("service-1"),
			}
		},
		GetCacheConfigFunc: func() *configv1.CacheConfig {
			return cacheConfig
		},
	}

	// Setup context
	ctx := context.Background()
	ctx = tool.NewContextWithTool(ctx, mockTool)

	// Key calculation: fmt.Sprintf("%s:%s", req.ToolName, req.ToolInputs)
	// "test-tool:{}"
	cacheKey := "test-tool:{}"

	// Test Case 1: Cache Miss, Set Cache
	mockRedis.ExpectGet(cacheKey).RedisNil()
	mockRedis.ExpectSet(cacheKey, "result", 1*time.Minute).SetVal("OK")

	res, err := m.Execute(ctx, req, func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "result", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "result", res)
	assert.NoError(t, mockRedis.ExpectationsWereMet())

	// Test Case 2: Cache Hit
	// Note: We reuse the same middleware instance, so the client/store is already cached.
	// But mockRedis is the underlying client, so we can set new expectations on it.
	mockRedis.ExpectGet(cacheKey).SetVal("cached-result")

	res, err = m.Execute(ctx, req, func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "should-not-be-called", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "cached-result", res)
	assert.NoError(t, mockRedis.ExpectationsWereMet())
}
