// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestCachingMiddleware_Redis_Write(t *testing.T) {
	db, mock := redismock.NewClientMock()

	// Override client creator
	middleware.SetRedisClientCreatorForTests(func(opts *redis.Options) *redis.Client {
		return db
	})
	defer middleware.SetRedisClientCreatorForTests(redis.NewClient)

	// Setup
	tm := &mockToolManager{}
	cacheMiddleware := middleware.NewCachingMiddleware(tm)

	ttl := 1 * time.Hour

	// Use builder for Opaque message
	redisConfig := bus.RedisBus_builder{
		Address: proto.String("localhost:6379"),
	}.Build()

	storageRedis := configv1.CacheConfig_STORAGE_REDIS

	testTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("redis-tool"),
			ServiceId: proto.String("redis-service"),
		}.Build(),
		cacheConfig: configv1.CacheConfig_builder{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(ttl),
			Storage:   &storageRedis,
			Redis:     redisConfig,
		}.Build(),
	}

	inputMap := map[string]any{"arg": "val"}
	inputJSON, _ := json.Marshal(inputMap)

	req := &tool.ExecutionRequest{
		ToolName:   "redis-service.redis-tool",
		ToolInputs: json.RawMessage(inputJSON),
	}
	ctx := tool.NewContextWithTool(context.Background(), testTool)

	nextFunc := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		t, _ := tool.GetFromContext(ctx)
		return t.Execute(ctx, req)
	}

	// Canonical Key
	expectedKey := `redis-service.redis-tool:{"arg":"val"}`

	// Expect GET (Cache Miss)
	mock.ExpectGet(expectedKey).RedisNil()

	// Expect SET (Cache Write)
	mock.ExpectSet(expectedKey, "success", ttl).SetVal("OK")

	// Execute
	res, err := cacheMiddleware.Execute(ctx, req, nextFunc)
	require.NoError(t, err)
	assert.Equal(t, successResult, res)
	assert.Equal(t, 1, testTool.executeCount)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
